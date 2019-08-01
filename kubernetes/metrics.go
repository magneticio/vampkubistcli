package kubeclient

import (
	"encoding/json"
	"errors"
	"github.com/magneticio/vampkubistcli/logging"
	//	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
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
			Name              string    `json:"name"`
			Namespace         string    `json:"namespace"`
			SelfLink          string    `json:"selfLink"`
			CreationTimestamp time.Time `json:"creationTimestamp"`
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

// Metrics returns list of metrics for a given namespace
func GetMetrics(configPath string, namespace string, pods *PodMetricsList) error {
	clientset, _, err := getLocalKubeClient(configPath)
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
	ProcessMetrics(&pods)
	return nil
}

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
		f := v.Field(i)
		switch f.Kind() {
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
					ProcessField(&v, "CPUf", cpu, err)
				case "memory":
					mem, err := ConvertMemory(f.String())
					ProcessField(&v, "MemoryF", mem, err)
				}
			}

		}
	}
}

func ProcessField(stract *reflect.Value, fieldName string, val float64, err error) {
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
