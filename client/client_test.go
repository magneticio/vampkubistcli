package client_test

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/magneticio/vampkubistcli/client"
	"github.com/magneticio/vampkubistcli/logging"
	"gopkg.in/resty.v1"
)

func createTestServer(t *testing.T, fn func(w http.ResponseWriter, r *http.Request) bool) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Method: %v", r.Method)
		t.Logf("Path: %v", r.URL.Path)
		if !fn(w, r) {
			t.Logf("Unhandled Path: %v", r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
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
	ts := createTestServer(t, func(w http.ResponseWriter, r *http.Request) bool {
		if r.Method == resty.MethodPost && r.URL.Path == "/oauth/access_token" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"token_type": "Bearer","access_token": "Test-Access-Token","expires_in": 3599,"refresh_token": "Test-Refresh-Token"}`))
			return true
		}
		return false
	})
	defer ts.Close()

	Cert := ""
	Username := "test"
	Password := "pass"
	Version := "v1"
	restClient := client.NewRestClient(ts.URL, Version, logging.Verbose, Cert)
	err := restClient.Login(Username, Password)

	assertError(t, err)
	assertEqual(t, "Test-Refresh-Token", restClient.RefreshToken())
}

func TestClientListErrorMessage(t *testing.T) {
	ts := createTestServer(t, func(w http.ResponseWriter, r *http.Request) bool {
		if r.Method == resty.MethodGet && r.URL.Path == "/api/v1/examples/list" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"message": "ERROR MESSAGE"}`))
			return true
		}
		return false
	})
	defer ts.Close()
	Cert := ""
	Version := "v1"
	Verbose := false
	restClient := client.NewRestClient(ts.URL, Version, Verbose, Cert)
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

func TestFallbacktoRefresh(t *testing.T) {
	refreshCalled := false
	ts := createTestServer(t, func(w http.ResponseWriter, r *http.Request) bool {
		if r.Method == resty.MethodPost {
			switch r.URL.Path {
			case "/api/v1/projects":
				if refreshCalled {
					w.WriteHeader(http.StatusOK)
				} else {
					w.WriteHeader(http.StatusUnauthorized)
				}
				return true
			case "/oauth/access_token":
				w.WriteHeader(http.StatusOK)
				refreshCalled = true
				return true
			}
		}
		return false
	})
	defer ts.Close()
	restClient := client.NewRestClient(ts.URL, "v1", false, "")
	values := make(map[string]string)
	values["cluster"] = "cluster"
	values["virtual_cluster"] = "virtualcluster"
	values["application"] = "application"
	result, err := restClient.Create("project", "test_project", `{"metadata": {"key1": "value1"}}`, "json", values)
	assertEqual(t, true, result)
	assertEqual(t, nil, err)
}
