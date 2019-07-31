// Copyright © 2018 Developer developer@vamp.io
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/gorilla/websocket"
	"github.com/magneticio/vampkubistcli/logging"
	"github.com/magneticio/vampkubistcli/models"
	"gopkg.in/resty.v1"
)

/*
For user friendliness, a resource map is used to map resource types
*/
// GO doesn't allow const maps so this is a var
var resourceMap map[string]string = map[string]string{
	"project":          "project",
	"projects":         "project",
	"cluster":          "cluster",
	"clusters":         "cluster",
	"virtual_cluster":  "virtual_cluster",
	"virtual_clusters": "virtual_cluster",
	"virtualcluster":   "virtual_cluster",
	"virtualclusters":  "virtual_cluster",
	"gateway":          "gateway",
	"gateways":         "gateway",
	"vamp_service":     "vamp_service",
	"vamp_services":    "vamp_service",
	"vampservice":      "vamp_service",
	"vampservices":     "vamp_service",
	"application":      "application",
	"applications":     "application",
	"destination":      "destination",
	"destinations":     "destination",
	"canary_release":   "canary_release",
	"canary_releases":  "canary_release",
	"canaryrelease":    "canary_release",
	"canaryreleases":   "canary_release",
	"service":          "service",
	"services":         "service",
	"serviceentry":     "service_entry",
	"serviceentries":   "service_entry",
	"service_entries":  "service_entry",
	"service_entry":    "service_entry",
	"deployment":       "deployment",
	"deployments":      "deployment",
	"role":             "role",
	"roles":            "role",
	"user":             "user",
	"users":            "user",
	"permission":       "permission",
	"permissions":      "permission",
	"experiment":       "experiment",
	"experiments":      "experiment",
}

type RestClient struct {
	URL            string
	Version        string
	Username       string
	Password       string
	Certs          string
	RefreshToken   string
	ExpirationTime int64
	TokenStore     *TokenStore
}

type successResponse struct {
	/* variables */
	Message string
}

type errorResponse struct {
	/* variables */
	Message string
}

type authSuccess struct {
	/* variables */
	TokenType    string `json:"token_type"`
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

const defaultVersion = "v1"

const defaultTimeout = 30 * time.Second

func NewRestClient(url string, token string, version string, isVerbose bool, cert string, tokenStore *TokenStore) *RestClient {
	if url == "" {
		logging.Error("URL can not be empty, check your configuration")
		return nil
	}
	url = strings.TrimRight(url, "/") // Url should end without a /
	resty.SetDebug(isVerbose)
	if version == "" {
		logging.Info("Using Default Version for client: %s\n", defaultVersion)
		version = defaultVersion
	}
	if cert != "" {
		// Create our Temp File:  This will create a filename like /tmp/prefix-123456
		// We can use a pattern of "pre-*.txt" to get an extension like: /tmp/pre-123456.txt
		tmpFile, err := ioutil.TempFile(os.TempDir(), "vamp-")
		if err != nil {
			logging.Error("Can not create temporary file: %v\n", err)
			return nil
		}
		// Remember to clean up the file afterwards
		defer os.Remove(tmpFile.Name())
		err_write := ioutil.WriteFile(tmpFile.Name(), []byte(cert), 0644)
		if err_write != nil {
			logging.Error("Can not write temporary file: %v\n", err_write)
			return nil
		}
		// fmt.Printf("load cert from file: %v\n", tmpFile.Name())
		resty.SetRootCertificate(tmpFile.Name())
	}
	retryCount := 5
	// Set retry wait times that do not intersect with default ones
	retryWaitTime := time.Duration(3) * time.Second
	retryMaxWaitTime := time.Duration(9) * time.Second
	resty.
		SetRetryCount(retryCount).
		SetRetryWaitTime(retryWaitTime).
		SetRetryMaxWaitTime(retryMaxWaitTime).
		AddRetryCondition(
			func(r *resty.Response) (bool, error) {
				return r.StatusCode() >= http.StatusBadRequest, nil
			},
		)
	// default timeout of golang is very long
	resty.SetTimeout(defaultTimeout)
	logging.Info("Rest client base url: %v\n", url)

	var tokenStoreImp TokenStore
	if tokenStore != nil {
		tokenStoreImp = *tokenStore
	} else {
		tokenStoreImp = &InMemoryTokenStore{}
	}
	/*
			  This a bit tricky, user can start with a access token or a refresh token
			  If it is an access token,
			  user has 30 seconds to use it starting from client creation.
			  If it is an refresh token,
			  access token will fail and user will get a new access token
		    TODO: this feature doesn't work right now.
		    tokenStoreImp.Store(token, time.Now().Unix()+30)
	*/
	return &RestClient{
		URL:          url,
		RefreshToken: token,
		Version:      version,
		Certs:        cert,
		TokenStore:   &tokenStoreImp,
	}
}

/*
Function to get Success message
*/
func getSuccessMessage(resp *resty.Response) string {
	return resp.Result().(*successResponse).Message
}

/*
Return the error object with message to be returned
It can be used after checking with IsError
*/
func getError(resp *resty.Response) error {
	message := resp.Error().(*errorResponse).Message
	if message == "" {
		message = string(resp.Body())
	}

	var responseBody models.ErrorResponse

	json.Unmarshal([]byte(resp.Body()), &responseBody)

	if responseBody.ValidationOutcome != nil {
		for _, element := range responseBody.ValidationOutcome {
			message = message + "\n\t- " + element.Error
		}
	}

	return errors.New(message)
}

/*
This is added for user friendliness.
If a user uses a plural name or misses an underscore,
api will still able to work
*/
func ResourceTypeConversion(resource string) string {
	// everything is lower case in the api
	// only _ is used in rest api
	resourceString := strings.Replace(strings.ToLower(resource), "-", "_", -1)
	if val, ok := resourceMap[resourceString]; ok {
		return val
	}

	return resourceString

}

func (s *RestClient) getAccessToken() string {
	activeToken := ""
	(*s.TokenStore).RemoveExpired()
	latest := time.Now().Unix()
	for token, timeout := range (*s.TokenStore).Tokens() {
		// it should have at least 10 seconds to expire
		if time.Now().Unix() < timeout-10 {
			if timeout > latest {
				activeToken = token
			}
		}
	}
	if activeToken == "" {
		logging.Info("Access token is expired - refreshing...")
		token, error := s.RefreshTokens()
		if error == nil {
			return token
		}
		logging.Error("Refresh Token Error: %v\n", error)
	}
	return activeToken
}

func (s *RestClient) parseTokenResponse(resp *authSuccess) string {
	token := resp.AccessToken
	s.RefreshToken = resp.RefreshToken
	expiresIn := resp.ExpiresIn
	if expiresIn != 0 {
		expirationTime := time.Now().Unix() + expiresIn
		(*s.TokenStore).Store(token, expirationTime)
	}
	return s.RefreshToken
}

func (s *RestClient) auth(body string) (string, error) {
	url := s.URL + "/oauth/access_token"
	resp, err := resty.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded; charset=utf-8").
		SetHeader("Accept", "application/json").
		SetBody([]byte(body)).
		SetResult(&authSuccess{}).
		SetError(&errorResponse{}).
		Post(url)

	if err != nil {
		return "", err
	}

	if resp != nil {
		if resp.IsError() {
			return "", errors.New(string(resp.Body()))
		}
		(*s.TokenStore).Clean()
		refreshToken := s.parseTokenResponse(resp.Result().(*authSuccess))
		return refreshToken, nil
	}

	return "", errors.New("Authentication failed")
}

func (s *RestClient) Login(username string, password string) (string, error) {
	s.Username = username
	s.Password = password
	body := "username=" + username + "&password=" + password + "&client_id=frontend&client_secret=&grant_type=password"
	return s.auth(body)
}

func (s *RestClient) RefreshTokens() (string, error) {
	body := "client_id=frontend&client_secret=&grant_type=refresh_token&refresh_token=" + s.RefreshToken
	return s.auth(body)
}

func (s *RestClient) fallbackToRefreshToken(f func() (*resty.Response, error)) (*resty.Response, error) {
	resp, err := f()
	if err == nil && resp.IsError() {
		if resp.StatusCode() == http.StatusUnauthorized {
			logging.Info("Got StatusUnauthorized: ( %v ), refreshing token...", http.StatusUnauthorized)
			if _, refreshTokensErr := s.RefreshTokens(); refreshTokensErr != nil {
				logging.Error("Cannot refresh token - %v\n", refreshTokensErr)
				return nil, refreshTokensErr
			}
			return f()
		}
		return resp, getError(resp)
	}
	return resp, err
}

func getUrlForResource(base string, version string, resourceName string, subCommand string, name string, values map[string]string) (string, error) {
	resourceName = ResourceTypeConversion(resourceName)
	subPath := ""
	namedParameter := ""
	if subCommand != "" {
		subPath = "/" + subCommand
	}
	if name != "" {
		namedParameter = "&" + resourceName + "_name=" + name
	}

	project := values["project"]

	cluster := values["cluster"]

	virtualCluster := values["virtual_cluster"]

	switch resourceName {
	case "project":
		return base + "/api/" + version + "/" + "projects" + subPath + "?time=-1" + namedParameter, nil
	case "cluster":
		url := base + "/api/" + version + "/" + "clusters" + subPath +
			"?" + "project_name=" + project +
			namedParameter
		return url, nil
	case "virtual_cluster":
		url := base + "/api/" + version + "/" + "virtual-clusters" + subPath +
			"?" + "project_name=" + project +
			"&" + "cluster_name=" + cluster +
			namedParameter
		return url, nil
	case "virtual_service":
		url := base + "/api/" + version + "/" + "virtual-services" + subPath +
			"?" + "project_name=" + project +
			"&" + "cluster_name=" + cluster +
			"&" + "virtual_cluster_name=" + virtualCluster +
			namedParameter
		return url, nil
	}
	var queryParams = ""

	for key, value := range values {
		if value != "" {
			queryParams = queryParams + "&" + key + "_name=" + value
		}
	}

	timestamp := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	url := base + "/api/" + version + "/" + fixResourceName(resourceName) + subPath +
		"?" + "t=" + timestamp +
		queryParams +
		namedParameter
	return url, nil
}

func fixResourceName(resourceName string) string {

	if resourceMap[resourceName] == "service_entry" {
		return "service-entries"
	} else {
		return strings.Replace(resourceName, "_", "-", -1) + "s"
	}

}

func (s *RestClient) Create(resourceName string, name string, source string, sourceType string, values map[string]string) (bool, error) {
	return (*s).Apply(resourceName, name, source, sourceType, values, false)
}

func (s *RestClient) Update(resourceName string, name string, source string, sourceType string, values map[string]string) (bool, error) {
	return (*s).Apply(resourceName, name, source, sourceType, values, true)
}

func (s *RestClient) Apply(resourceName string, name string, source string, sourceType string, values map[string]string, update bool) (bool, error) {

	if sourceType == "yaml" {
		json, err := yaml.YAMLToJSON([]byte(source))
		if err != nil {
			return false, err
		}
		source = string(json)
	}

	body := []byte(source)

	version, jsonErr := getVersionFromResource(body)
	if jsonErr != nil {
		return false, jsonErr
	}

	if version == "" {
		version = s.Version
	}

	url, _ := getUrlForResource(s.URL, version, resourceName, "", name, values)
	logging.Info("Requesting url: %v\n", url)
	var resp *resty.Response
	var err error
	if update {
		resp, err = s.fallbackToRefreshToken(func() (*resty.Response, error) {
			return resty.R().
				SetHeader("Content-Type", "application/json").
				SetHeader("Accept", "application/json").
				SetAuthToken(s.getAccessToken()).
				SetBody(body).
				SetResult(&successResponse{}).
				SetError(&errorResponse{}).
				Put(url)
		})
	} else {
		resp, err = s.fallbackToRefreshToken(func() (*resty.Response, error) {
			return resty.R().
				SetHeader("Content-Type", "application/json").
				SetHeader("Accept", "application/json").
				SetAuthToken(s.getAccessToken()).
				SetBody(body).
				SetResult(&successResponse{}).
				SetError(&errorResponse{}).
				Post(url)
		})
	}

	if err != nil {
		return false, err
	}

	if resp.IsError() {
		return false, getError(resp)
	}
	return true, nil

}

func (s *RestClient) Delete(resourceName string, name string, values map[string]string) (bool, error) {
	url, _ := getUrlForResource(s.URL, s.Version, resourceName, "", name, values)
	resp, err := s.fallbackToRefreshToken(func() (*resty.Response, error) {
		return resty.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Accept", "application/json").
			SetAuthToken(s.getAccessToken()).
			SetResult(&successResponse{}).
			SetError(&errorResponse{}).
			Delete(url)
	})

	if err != nil {
		return false, err
	}

	if resp.IsError() {
		return false, getError(resp)
	}
	return true, nil

}

func (s *RestClient) GetSpec(resourceName string, name string, outputFormat string, values map[string]string) (string, error) {
	url, _ := getUrlForResource(s.URL, s.Version, resourceName, "", name, values)

	resp, getResourceError := s.fallbackToRefreshToken(func() (*resty.Response, error) {
		return resty.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Accept", "application/json").
			SetAuthToken(s.getAccessToken()).
			SetError(&errorResponse{}).
			Get(url)
	})

	if getResourceError != nil {
		return "", getResourceError
	}

	if resp.IsError() {
		return "", getError(resp)
	}

	var withSpec models.WithSpecification
	unmarshalError := json.Unmarshal(resp.Body(), &withSpec)
	if unmarshalError != nil {
		return "", unmarshalError
	}

	specification, marshallSpecError := json.Marshal(withSpec.Specification)
	if marshallSpecError != nil {
		return "", marshallSpecError
	}

	source := ""
	if outputFormat == "yaml" {
		yaml, err_2 := yaml.JSONToYAML(specification)
		if err_2 != nil {
			return "", err_2
		}
		source = string(yaml)
	} else {
		var prettyJSON bytes.Buffer
		error := json.Indent(&prettyJSON, specification, "", "    ")
		if error != nil {
			return "", error
		}
		source = string(prettyJSON.Bytes())
	}
	return source, nil

}

func (s *RestClient) Get(resourceName string, name string, outputFormat string, values map[string]string) (string, error) {
	url, _ := getUrlForResource(s.URL, s.Version, resourceName, "", name, values)

	resp, err := s.fallbackToRefreshToken(func() (*resty.Response, error) {
		return resty.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Accept", "application/json").
			SetAuthToken(s.getAccessToken()).
			SetError(&errorResponse{}).
			Get(url)
	})

	if err != nil {
		return "", err
	}

	if resp.IsError() {
		return "", getError(resp)
	}
	source := ""
	if outputFormat == "yaml" {
		yaml, err_2 := yaml.JSONToYAML(resp.Body())
		if err_2 != nil {
			return "", err_2
		}
		source = string(yaml)
	} else {
		var prettyJSON bytes.Buffer
		error := json.Indent(&prettyJSON, resp.Body(), "", "    ")
		if error != nil {
			return "", error
		}
		source = string(prettyJSON.Bytes())
	}
	return source, nil

}

func (s *RestClient) List(resourceName string, outputFormat string, values map[string]string, simple bool) (string, error) {
	url, _ := getUrlForResource(s.URL, s.Version, resourceName, "list", "", values)

	resp, err := s.fallbackToRefreshToken(func() (*resty.Response, error) {
		return resty.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Accept", "application/json").
			SetAuthToken(s.getAccessToken()).
			SetError(&errorResponse{}).
			Get(url)
	})

	if err != nil {
		return "", err
	}

	if resp.IsError() {
		return "", getError(resp)
	}

	responseBody := resp.Body()
	if simple {
		var r []models.Named
		err := json.Unmarshal([]byte(responseBody), &r)
		if err != nil {
			return "", errors.New(string(responseBody))
		}
		// Array conversion is done to show only names
		arr := make([]string, len(r))
		for i, named := range r {
			arr[i] = named.Name
		}
		responseBody, err = json.Marshal(arr)
		if err != nil {
			return "", err
		}
	}

	source := ""
	if outputFormat == "yaml" {
		yaml, err_2 := yaml.JSONToYAML(responseBody)
		if err_2 != nil {
			return "", err_2
		}
		source = string(yaml)
	} else {
		var prettyJSON bytes.Buffer
		error := json.Indent(&prettyJSON, responseBody, "", "    ")
		if error != nil {
			return "", error
		}
		source = string(prettyJSON.Bytes())
	}
	return source, nil

}

func (s *RestClient) UpdateUserPermission(username string, permission string, values map[string]string) (bool, error) {
	url, _ := getUrlForResource(s.URL, s.Version, "user-access-permission", "", "", values)
	url += "&user_name=" + username

	permissionBody := models.Permission{
		Read:       strings.Contains(permission, "r"),
		Write:      strings.Contains(permission, "w"),
		Delete:     strings.Contains(permission, "d"),
		EditAccess: strings.Contains(permission, "a"),
	}

	resp, err := s.fallbackToRefreshToken(func() (*resty.Response, error) {
		return resty.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Accept", "application/json").
			SetAuthToken(s.getAccessToken()).
			SetBody(permissionBody).
			SetResult(&successResponse{}).
			SetError(&errorResponse{}).
			Post(url)
	})

	if err != nil {
		return false, err
	}

	if resp.IsError() {
		return false, getError(resp)
	}
	return true, nil

}

func (s *RestClient) RemovePermissionFromUser(username string, values map[string]string) (bool, error) {
	url, _ := getUrlForResource(s.URL, s.Version, "user-access-permission", "", "", values)
	url += "&user_name=" + username
	resp, err := s.fallbackToRefreshToken(func() (*resty.Response, error) {
		return resty.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Accept", "application/json").
			SetAuthToken(s.getAccessToken()).
			SetResult(&authSuccess{}).
			SetError(&errorResponse{}).
			Delete(url)
	})

	if err != nil {
		return false, err
	}

	if resp.IsError() {
		return false, getError(resp)
	}
	return true, nil

}

func (s *RestClient) AddRoleToUser(username string, rolename string, values map[string]string) (bool, error) {
	url, _ := getUrlForResource(s.URL, s.Version, "user-access-role", "", "", values)
	url += "&user_name=" + username + "&role_name=" + rolename
	resp, err := s.fallbackToRefreshToken(func() (*resty.Response, error) {
		return resty.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Accept", "application/json").
			SetAuthToken(s.getAccessToken()).
			SetResult(&successResponse{}).
			SetError(&errorResponse{}).
			Post(url)
	})

	if err != nil {
		return false, err
	}

	if resp.IsError() {
		return false, getError(resp)
	}
	return true, nil

}

func (s *RestClient) RemoveRoleFromUser(username string, rolename string, values map[string]string) (bool, error) {
	url, _ := getUrlForResource(s.URL, s.Version, "user-access-role", "", "", values)
	url += "&user_name=" + username + "&role_name=" + rolename
	resp, err := s.fallbackToRefreshToken(func() (*resty.Response, error) {
		return resty.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Accept", "application/json").
			SetAuthToken(s.getAccessToken()).
			SetResult(&authSuccess{}).
			SetError(&errorResponse{}).
			Delete(url)
	})

	if err != nil {
		return false, err
	}

	if resp.IsError() {
		return false, getError(resp)
	}
	return true, nil

}

/*
Ping is different from other calls
It just runs a get to the root folder and doesn't check anything
*/
func (s *RestClient) Ping() (bool, error) {
	url := s.URL + "/"
	resty.SetTimeout(5 * time.Second)
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "text/plain, application/json").
		// Should be reachable without a token SetAuthToken(s.Token).
		Get(url)

	if err != nil {
		return false, err
	}

	if resp.IsError() {
		return false, errors.New(string(resp.Body()))
	}
	return true, nil

}

func getVersionFromResource(source []byte) (string, error) {

	var Versioned models.Versioned

	jsonErr := json.Unmarshal(source, &Versioned)
	if jsonErr != nil {
		return "", jsonErr
	}

	return Versioned.Version, nil

}

func (s *RestClient) ReadNotifications(notifications chan<- models.Notification) error {

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u, urlParseError := url.Parse(s.URL + "/api/" + s.Version + "/notifications?access_token=" + s.getAccessToken())
	if urlParseError != nil {
		return urlParseError
	}
	u.Scheme = "wss"
	logging.Info("connecting to %s", u.String())
	dialer := websocket.DefaultDialer
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(s.Certs))
	if !ok {
		return errors.New("failed to parse root certificate")
	}
	dialer.TLSClientConfig = &tls.Config{
		RootCAs: roots,
	}
	c, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		logging.Error("dial: %v", err)
		return err
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			// {"text":"Initializing cluster cluster1"}
			var notification models.Notification
			err := c.ReadJSON(&notification)
			if err != nil {
				logging.Info("read: %v", err)
				// return err
			}
			notifications <- notification
			logging.Info("recv: %s", notification.Text)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return nil
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				logging.Info("write: %v", err)
				return err
			}
		case <-interrupt:
			logging.Info("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				logging.Info("write close: %v", err)
				return err
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return nil
		}
	}
}

func (s *RestClient) SendExperimentMetric(experimentName string, metricName string, experimentMetric *models.ExperimentMetric, values map[string]string) error {
	url, _ := getUrlForResource(s.URL, s.Version, "experiments", "metrics", "", values)
	url += "&experiment_name=" + experimentName + "&metric_name=" + metricName
	strJson, jsonMarshalError := json.Marshal(*experimentMetric)
	if jsonMarshalError != nil {
		return jsonMarshalError
	}
	body := strJson
	resp, err := s.fallbackToRefreshToken(func() (*resty.Response, error) {
		return resty.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Accept", "application/json").
			SetAuthToken(s.getAccessToken()).
			SetBody(body).
			SetResult(&successResponse{}).
			SetError(&errorResponse{}).
			Put(url)
	})
	if err != nil {
		return err
	}

	if resp.IsError() {
		return errors.New(string(resp.Body()))
	}
	return nil
}

// GetSubsetMap returns a map for easy conversion of labels to subsets
func (s *RestClient) GetSubsetMap(values map[string]string) (*models.DestinationsSubsetsMap, error) {
	url, _ := getUrlForResource(s.URL, s.Version, "destination", "subsets/map", "", values)

	resp, err := s.fallbackToRefreshToken(func() (*resty.Response, error) {
		return resty.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Accept", "application/json").
			SetAuthToken(s.getAccessToken()).
			SetError(&errorResponse{}).
			Get(url)
	})

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, getError(resp)
	}

	responseBody := resp.Body()

	var destinationsSubsetsMap models.DestinationsSubsetsMap
	unmarshalError := json.Unmarshal(responseBody, &destinationsSubsetsMap)
	if unmarshalError != nil {
		return nil, unmarshalError
	}

	return &destinationsSubsetsMap, nil
}
