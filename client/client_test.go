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
				_, _ = w.Write([]byte(`{"token_type": "Bearer","access_token": "Test-Access-Token","expires_in": 3599,"refresh_token": "Test-Refresh-Token"}`))
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
	restClient := client.NewRestClient(ts.URL, Token, Version, logging.Verbose, Cert)
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
			}
		}
	})
	defer ts.Close()
	Token := "Test-Token"
	Cert := ""
	Version := "v1"
	Verbose := false
	restClient := client.NewRestClient(ts.URL, Token, Version, Verbose, Cert)
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
