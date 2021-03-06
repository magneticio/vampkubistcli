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

func TestGetSimpleMetrics(t *testing.T) {
	ts := CreateMockedK8s(t, "metrics_test.json")
	defer ts.Close()

	metrics, err := kubeclient.GetSimpleMetrics("", "vamp-system")

	if err != nil {
		t.Errorf("GetSimpleMetrics returned error: %v", err)
	}

	if len(metrics) == 0 {
		t.Error("GetSimpleMetrics should return data")
	}

	for _, pod := range metrics {
		if pod.Name == "" {
			t.Errorf("Name should not be empty in %v", pod)
		}
		if len(pod.Labels) == 0 {
			t.Errorf("Labels should not be empty in %v", pod)
		}
		for _, cnt := range pod.ContainersMetrics {
			if cnt.CPU == 0 {
				t.Errorf("CPU should not be zero in %v", cnt)
			}
			if cnt.Memory == 0 {
				t.Errorf("Memory should not be zero in %v", cnt)
			}
		}
	}

	if metrics[2].Name != "mongo-0" {
		t.Errorf("Pod name isn't correct in %v", metrics[2])
	}
	if len(metrics[2].ContainersMetrics) != 2 {
		t.Errorf("Containers number isn't correct in %v", metrics[2])
	}
	if metrics[2].ContainersMetrics[0].Name != "mongo" {
		t.Errorf("Container name isn't correct in %v", metrics[2])
	}
	if metrics[2].ContainersMetrics[0].CPU != 7261378e-9 {
		t.Errorf("CPU value isn't correct in %v", metrics[2])
	}
	if metrics[2].ContainersMetrics[1].Memory != 67900*1024 {
		t.Errorf("Memory value isn't correct in %v", metrics[2])
	}
	t.Logf("--pods: \n%v", metrics)
}

func TestGetAverageMetrics(t *testing.T) {
	ts := CreateMockedK8s(t, "metrics_test.json")
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

	found := map[string]bool{"vamp-6dc7f8cd87-47kdw": false, "mongo-2": false}
	for _, m := range metrics {
		if m.Name == "vamp-6dc7f8cd87-47kdw" {
			found["vamp-6dc7f8cd87-47kdw"] = true
			expectedCPU := 24946926e-9
			if m.CPU != expectedCPU {
				t.Errorf("Expected %v, got %v", expectedCPU, m.CPU)
			}
			expectedMem := float64(308068) * 1024
			if m.Memory != expectedMem {
				t.Errorf("Expected %v, got %v", expectedMem, m.Memory)
			}
		}
		if m.Name == "mongo-2" {
			found["mongo-2"] = true
			expectedCPU := (float64(6940819e-9) + 2563134e-9) / 2
			if m.CPU != expectedCPU {
				t.Errorf("Expected %v, got %v", expectedCPU, m.CPU)
			}
			expectedMem := float64(59324+67568) * 1024 / 2
			if m.Memory != expectedMem {
				t.Errorf("Expected %v, got %v", expectedMem, m.Memory)
			}
		}

	}

	for k, v := range found {
		if v != true {
			t.Errorf("There is no %v in metrics json", k)
		}
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
