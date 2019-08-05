package kubeclient_test

import (
	kubeclient "github.com/magneticio/vampkubistcli/kubernetes"
	"io/ioutil"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func createTestServer(fn func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(fn))
}

type k8sClientProviderMock struct {
	Host string
}

func (mock k8sClientProviderMock) Get(configPath string) (*kubernetes.Clientset, error) {
	cfg := rest.Config{
		Host: mock.Host,
	}
	return kubernetes.NewForConfig(&cfg)
}

func CreateMockedK8s(t *testing.T, metricsFileName string) *httptest.Server {
	metricsJS, err := ioutil.ReadFile(metricsFileName)
	if err != nil {
		t.Errorf("Cannot read metrics json file - %v", err)
	}

	podJS, err := ioutil.ReadFile("pod_test.json")
	if err != nil {
		t.Errorf("Cannot read pod json file - %v", err)
	}

	ts := createTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Method: %v", r.Method)
		t.Logf("Path: %v", r.URL.Path)
		switch {
		case r.URL.Path == "/apis/metrics.k8s.io/v1beta1/namespaces/vamp-system/pods":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(metricsJS)
		case strings.HasPrefix(r.URL.Path, "/api/v1/namespaces/vamp-system/pods/"):
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(podJS)
		default:
		}
	})

	kubeclient.K8sClient = k8sClientProviderMock{Host: ts.URL}

	return ts
}

func TestGetProcessedMetrics(t *testing.T) {
	ts := CreateMockedK8s(t, "metrics_test.json")
	defer ts.Close()

	var pods kubeclient.PodMetricsList

	if err := kubeclient.GetProcessedMetrics("", "vamp-system", &pods); err != nil {
		t.Errorf("GetProcessedMetrics returned error: %v", err)
	}

	if len(pods.Items) == 0 {
		t.Error("GetProcessedMetrics should return data")
	}

	for _, item := range pods.Items {
		for _, cnt := range item.Containers {
			if cnt.Usage.CPUf == 0 {
				t.Errorf("CPUf should not be zero in %v", cnt)
			}
			if cnt.Usage.MemoryF == 0 {
				t.Errorf("MemoryF should not be zero in %v", cnt)
			}
		}
	}
	t.Logf("--pods: \n%v", pods)
}

func TestGetAverageMetrics(t *testing.T) {
	ts := CreateMockedK8s(t, "metrics_test.json")
	defer ts.Close()

	switch metrics, err := kubeclient.GetAverageMetrics("", "vamp-system"); {
	case err != nil:
		t.Errorf("GetAverageMetrics returned error: %v", err)
	case len(metrics) == 0:
		t.Error("GetAverageMetrics should return data")
	default:
		t.Logf("--metrics: \n%v", metrics)
	}
}

func TestGetAverageBadMetrics(t *testing.T) {
	ts := CreateMockedK8s(t, "metrics_test_bad.json")
	defer ts.Close()

	metrics, err := kubeclient.GetAverageMetrics("", "vamp-system")

	switch {
	case err != nil:
		t.Errorf("GetAverageMetrics returned error: %v", err)
	case len(metrics) == 0:
		t.Error("GetAverageMetrics should return data")
	default:
		t.Logf("--metrics: \n%v", metrics)
	}

	found := false
	badPodName := "mongo-0"
	for i := range metrics {
		if metrics[i].Name == badPodName {
			found = true
			if metrics[i].CPU != 0 {
				t.Errorf("CPU of %v should be 0 instead of %v", metrics[i].Name, metrics[i].CPU)
			}
			memExpectedVal := float64(67900) * 1024
			if metrics[i].Memory != memExpectedVal {
				t.Errorf("Memory of %v should be %v instead of %v", memExpectedVal, metrics[i].Name, metrics[i].Memory)
			}
		}
	}
	if !found {
		t.Errorf(`Metrics json file doesn't contain data of "%v" pod`, badPodName)
	}
}
