package kubeclient

import (
	"encoding/json"
	"errors"
	"github.com/magneticio/vampkubistcli/logging"
	//	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	//	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"time"
)

type PodMetricsList struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
	Metadata   struct {
		SelfLink string `json:"selfLink"`
	} `json:"metadata"`
	Items []struct {
		Metadata struct {
			Name              string            `json:"name"`
			Namespace         string            `json:"namespace"`
			SelfLink          string            `json:"selfLink"`
			CreationTimestamp time.Time         `json:"creationTimestamp"`
			Labels            map[string]string `json:"labels,omitempty" protobuf:"bytes,11,rep,name=labels"`
		} `json:"metadata"`
		Timestamp  time.Time `json:"timestamp"`
		Window     string    `json:"window"`
		Containers []struct {
			Name  string `json:"name"`
			Usage struct {
				CPU     string  `json:"cpu"`
				CPUf    float64 // field for storing processed CPU data
				Memory  string  `json:"memory"`
				MemoryF float64 // field for storing processed Memory data
			} `json:"usage"`
		} `json:"containers"`
	} `json:"items"`
}

// Interface for getting k8s client
type K8sClientProvider interface {
	Get(configPath string) (*kubernetes.Clientset, error)
}

type defK8sClient struct{}

// K8sClient provides k8s client that is used in metric methods that require interaction with k8s
var K8sClient K8sClientProvider = defK8sClient{}

// Get returns k8s client
func (defK8sClient) Get(configPath string) (*kubernetes.Clientset, error) {
	clientset, _, err := getLocalKubeClient(configPath)
	return clientset, err
}

// GetRawMetrics returns list of metrics for a given namespace
func GetRawMetrics(configPath string, namespace string, pods *PodMetricsList) error {
	clientset, err := K8sClient.Get(configPath)
	if err != nil {
		return err
	}
	data, err := clientset.RESTClient().Get().AbsPath("apis/metrics.k8s.io/v1beta1/namespaces/" + namespace + "/pods").DoRaw()
	if err != nil {
		return err
	}
	logging.Info("----raw json: %v", string(data))
	err = json.Unmarshal(data, &pods)
	if err != nil {
		return err
	}
	return nil
}

// GetProcessedMetrics extracts metrics from k8s using GetRawMetrics and then populates them with labels and
// float metrics that are converted from raw metrics' string values
func GetProcessedMetrics(configPath string, namespace string, pods *PodMetricsList) error {
	if err := GetRawMetrics(configPath, namespace, pods); err != nil {
		return err
	}

	if err := GetLabels(configPath, namespace, pods); err != nil {
		return err
	}

	ProcessMetrics(&pods)

	return nil
}

// GetAverageMetrics extract metrics from k8s using GetRawMetrics and then transforms them to
// new structure with labels and average CPU and memory per pod
func GetAverageMetrics(configPath string, namespace string) ([]PodAverageMetrics, error) {
	var pods PodMetricsList

	if err := GetRawMetrics(configPath, namespace, &pods); err != nil {
		return nil, err
	}

	if err := GetLabels(configPath, namespace, &pods); err != nil {
		return nil, err
	}

	return CalculateAverageMetrics(&pods)
}

// GetLabels populates pod metrics list with labels
func GetLabels(configPath string, namespace string, pods *PodMetricsList) error {
	clientset, err := K8sClient.Get(configPath)
	if err != nil {
		return err
	}

	for i := 0; i < len(pods.Items); i++ {
		if len(pods.Items[i].Metadata.Labels) == 0 {
			logging.Info("----getting labels for %v", pods.Items[i].Metadata.Name)
			pod, err := clientset.CoreV1().Pods(namespace).Get(pods.Items[i].Metadata.Name, metav1.GetOptions{})
			logging.Info("----got pod data: %v", pod)
			if err != nil {
				return err
			}
			pods.Items[i].Metadata.Labels = pod.Labels
		}
	}

	return nil
}

// PodAverageMetrics provides average CPU and memor as long as some pod's metadata
type PodAverageMetrics struct {
	Name   string
	Labels map[string]string
	CPU    float64
	Memory float64
}

func CalculateAverageMetrics(pods *PodMetricsList) ([]PodAverageMetrics, error) {
	var res = make([]PodAverageMetrics, len(pods.Items))
	for i := 0; i < len(pods.Items); i++ {
		res[i].Name = pods.Items[i].Metadata.Name
		res[i].Labels = pods.Items[i].Metadata.Labels
		var sumCPU, sumMem float64
		for j := 0; j < len(pods.Items[i].Containers); j++ {
			cpu, err := ConvertCPU(pods.Items[i].Containers[j].Usage.CPU)
			if err == nil {
				sumCPU += cpu
			} else {
				logging.Error("Conversion of CPU for %v failed - %v", pods.Items[i], err)
				return nil, err
			}
			mem, err := ConvertMemory(pods.Items[i].Containers[j].Usage.Memory)
			if err == nil {
				sumMem += mem
			} else {
				logging.Error("Conversion of Memory for %v failed - %v", pods.Items[i], err)
				return nil, err
			}
		}
		res[i].CPU = sumCPU / float64(len(pods.Items[i].Containers))
		res[i].Memory = sumMem / float64(len(pods.Items[i].Containers))
	}
	return res, nil
}

// ProcessMetrics converts string values for CPU and memory to float ones and stores them into dedicated fields
func ProcessMetrics(stract interface{}) {
	if reflect.ValueOf(stract).Kind() != reflect.Ptr {
		return
	}

	v := reflect.ValueOf(stract)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < v.NumField(); i++ {
		switch f := v.Field(i); f.Kind() {
		case reflect.Struct:
			ProcessMetrics(f.Addr().Interface())
		case reflect.Slice:
			for j := 0; j < f.Len(); j++ {
				ProcessMetrics(f.Index(j).Addr().Interface())
			}
		default:
			if json, ok := v.Type().Field(i).Tag.Lookup("json"); ok {
				logging.Info("----processing field with json tag %v", json)
				switch json {
				case "cpu":
					cpu, err := ConvertCPU(f.String())
					processField(&v, "CPUf", cpu, err)
				case "memory":
					mem, err := ConvertMemory(f.String())
					processField(&v, "MemoryF", mem, err)
				}
			}

		}
	}
}

func processField(stract *reflect.Value, fieldName string, val float64, err error) {
	if err == nil {
		f := stract.FieldByName(fieldName)
		if f.CanSet() {
			f.SetFloat(val)
		} else {
			logging.Error("cannot find field %v for storing processed data", fieldName)
		}
	} else {
		logging.Error("cannot process field %v - %v", fieldName, err)
	}
}

func ConvertCPU(cpu string) (float64, error) {
	units := map[string]string{
		"m": "e-3",
		"u": "e-6",
		"n": "e-9"}
	unit := cpu[len(cpu)-1:]
	if u, ok := units[unit]; ok {
		cpu = cpu[0:len(cpu)-1] + u
	}
	return strconv.ParseFloat(cpu, 64)
}

func ConvertMemory(mem string) (float64, error) {
	units := map[string]float64{
		"E":  1e18,
		"P":  1e15,
		"T":  1e12,
		"G":  1e9,
		"M":  1e6,
		"K":  1e3,
		"Ei": math.Pow(1024, 6),
		"Pi": math.Pow(1024, 5),
		"Ti": math.Pow(1024, 4),
		"Gi": math.Pow(1024, 3),
		"Mi": math.Pow(1024, 2),
		"Ki": 1024,
	}
	re := regexp.MustCompile("([0-9]+)(.*)")
	res := re.FindAllStringSubmatch(mem, -1)
	if len(res) == 0 || len(res[0]) < 3 {
		return 0, errors.New("cannot parse memory")
	}
	f1, err := strconv.ParseFloat(res[0][1], 64)
	if err != nil {
		return 0, errors.New("cannot parse digit part of memory")
	}
	f2, ok := units[res[0][2]]
	if !ok {
		f2 = 1
	}
	return f1 * f2, nil
}
