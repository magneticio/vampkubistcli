package kubeclient_test

import (
	"encoding/json"
	kubeclient "github.com/magneticio/vampkubistcli/kubernetes"
	"io/ioutil"
	"testing"
)

func TestProcessMetrics(t *testing.T) {
	js, err := ioutil.ReadFile("metrics_test.json")
	if err != nil {
		t.Errorf("Cannot read metrics json file - %v", err)
	}
	var pods kubeclient.PodMetricsList
	err = json.Unmarshal(js, &pods)
	if err != nil {
		t.Errorf("Cannot unmarshal metrics json file - %v", err)
	}
	kubeclient.ProcessMetrics(&pods)
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
