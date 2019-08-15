package client_test

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/magneticio/vampkubistcli/client"
	"github.com/magneticio/vampkubistcli/logging"
	"github.com/magneticio/vampkubistcli/models"
	"gopkg.in/resty.v1"
)

func createTestServer(fn func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(fn))
}

func assertError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("Error occurred [%v]", err)
	}
}

func equal(expected, got interface{}) bool {
	return reflect.DeepEqual(expected, got)
}

func assertEqual(t *testing.T, e, g interface{}) (r bool) {
	if !equal(e, g) {
		t.Errorf("Expected [%v], got [%v]", e, g)
	}

	return
}

func TestClientAuthToken(t *testing.T) {
	ts := createTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Method: %v", r.Method)
		t.Logf("Path: %v", r.URL.Path)
		if r.Method == resty.MethodPost {
			switch r.URL.Path {
			case "/oauth/access_token":
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"token_type": "Bearer","access_token": "Test-Access-Token","expires_in": 3599,"refresh_token": "Test-Access-Token"}`))
			default:
				t.Logf("Unhandled Path: %v", r.URL.Path)
			}
		}
	})
	defer ts.Close()

	Token := ""
	Cert := ""
	Username := "test"
	Password := "pass"
	Version := "v1"
	restClient := client.NewRestClient(ts.URL, Token, Version, logging.Verbose, Cert, nil)
	token, err := restClient.Login(Username, Password)

	assertError(t, err)
	assertEqual(t, "Test-Access-Token", token)
}

func TestClientListErrorMessage(t *testing.T) {
	ts := createTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Method: %v", r.Method)
		t.Logf("Path: %v", r.URL.Path)
		if r.Method == resty.MethodGet {
			switch r.URL.Path {
			case "/api/v1/examples/list":
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"message": "ERROR MESSAGE"}`))
			default:
				t.Logf("Unhandled Path: %v", r.URL.Path)
				w.WriteHeader(http.StatusAccepted)
				_, _ = w.Write([]byte(`{"message": "This is unexpected"}`))
			}
		}
	})
	defer ts.Close()
	Token := "Test-Token"
	Cert := ""
	Version := "v1"
	Verbose := false
	restClient := client.NewRestClient(ts.URL, Token, Version, Verbose, Cert, nil)
	values := make(map[string]string)
	values["project"] = "project"
	values["cluster"] = "cluster"
	values["virtual_cluster"] = "virtualcluster"
	values["application"] = "application"
	Type := "example"
	OutputType := "json"
	Detailed := false
	result, err := restClient.List(Type, OutputType, values, !Detailed)
	assertEqual(t, "", result)
	assertEqual(t, "ERROR MESSAGE", err.Error())
}

func TestClientPushMetricsModel(t *testing.T) {
	ts := createTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Method: %v", r.Method)
		t.Logf("Path: %v", r.URL.Path)
		if r.Method == resty.MethodPut {
			switch r.URL.Path {
			case "/api/v1/metrics/values":
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusAccepted)
				_, _ = w.Write([]byte(`{"message": "Metric is being updated"}`))
			default:
				t.Logf("Unhandled Path: %v", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"message": "Url is incorrect"}`))
			}
		}
	})
	defer ts.Close()
	Token := "Test-Token"
	Cert := ""
	Version := "v1"
	Verbose := false
	restClient := client.NewRestClient(ts.URL, Token, Version, Verbose, Cert, nil)
	values := make(map[string]string)
	values["project"] = "project"
	values["cluster"] = "cluster"
	values["virtual_cluster"] = "virtualcluster"
	values["destination"] = "destination"
	values["experiment"] = "experiment"

	metric := models.MetricValue{
		Timestamp:         1,
		NumberOfElements:  10,
		StandardDeviation: 0.3,
		Average:           0.9,
		Median:            0.6,
		Sum:               9,
		Min:               0.2,
		Max:               1,
		Rate:              0.3,
		P999:              0.1,
		P99:               0.2,
		P95:               0.6,
		P75:               0.9,
	}

	result, err := restClient.PushMetricValue("latency", &metric, values)
	assertEqual(t, true, result)
	assertEqual(t, nil, err)
}
