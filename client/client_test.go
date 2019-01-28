package client_test

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/magneticio/vampkubistcli/client"
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
			case "/json-invalid":
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte("TestGet: Invalid JSON"))
			}
		}
	})
	defer ts.Close()

	Token := ""
	Cert := ""
	Username := "test"
	Password := "pass"
	Debug := false
	restClient := client.NewRestClient(ts.URL, Token, Debug, Cert)
	token, err := restClient.Login(Username, Password)

	assertError(t, err)
	assertEqual(t, "Test-Access-Token", token)
}
