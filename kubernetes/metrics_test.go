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

func CreateMockedK8s(t *testing.T) *httptest.Server {
	metricsJS, err := ioutil.ReadFile("metrics_test.json")
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
	ts := CreateMockedK8s(t)
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
	ts := CreateMockedK8s(t)
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
