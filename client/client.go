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
	"cmd"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
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
}

type restClient struct {
	url            string
	version        string
	username       string
	password       string
	token          string
	certs          string
	refreshToken   string
	expirationTime int64
	config         *ClientConfig
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

func NewRestClient(url string, version string, isVerbose bool, cert string) *restClient {
	if url == "" {
		log.Fatal("URL can not be empty, check your configuration")
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
			log.Fatal("Can not create temporary file", err)
		}
		// Remember to clean up the file afterwards
		defer os.Remove(tmpFile.Name())
		err_write := ioutil.WriteFile(tmpFile.Name(), []byte(cert), 0644)
		if err_write != nil {
			log.Fatal("Can not write temporary file", err_write)
		}
		// fmt.Printf("load cert from file: %v\n", tmpFile.Name())
		resty.SetRootCertificate(tmpFile.Name())
	}
	// default timeout of golang is very long
	resty.SetTimeout(defaultTimeout)
	logging.Info("Rest client base url: %v\n", url)
	client := &restClient{
		url:     url,
		version: version,
		certs:   cert,
	}
	return client
}

func ClientFromConfig(cfg *ClientConfig, isVerbose bool) *restClient {
	client := NewRestClient(cfg.Url, cfg.APIVersion, isVerbose, cfg.Cert)
	client.config = cfg
	client.token = cfg.AccessToken
	client.refreshToken = cfg.RefreshToken
	client.expirationTime = cfg.ExpirationTime
	return client
}

func (s *restClient) refreshTokenIfNeeded() {
	if (time.Now().Unix() + 60) > s.expirationTime {
		logging.Info("Access token is expired - refreshing...")
		if cfg.RefreshToken == "" {
			log.Fatal("Cannot refresh token - current refresh token is empty")
		}
		err := s.RefreshToken()
		if err != nil {
			log.Fatal("Cannot refresh token - ", err)
		}
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

func (s *restClient) updateConfig() {
	if s.config != nill {
		s.config.AccessToken = s.token
		s.config.RefreshToken = s.refreshToken
		s.config.ExpirationTime = s.expirationTime
		writeConfigError := WriteConfigFile()
		if writeConfigError != nil {
			log.Fatal("Cannot save updated refresh token to config")
		}
	}
}

func (s *restClient) parseTokenResponse(resp *authSuccess) {
	(*s).token = resp.AccessToken
	(*s).refreshToken = resp.RefreshToken
	expiresIn := resp.ExpiresIn
	if expiresIn != "" {
		if tm, err := strconv.Atoi(expiresIn); err == nil {
			logging.Error("Cannot convert expiresIn response field %s to int", expiresIn)
		} else {
			(*s).expirationTime = time.Now().Unix() + tm
		}
	}
	s.updateConfig()
}

func (s *restClient) auth(body string) error {
	url := (*s).url + "/oauth/access_token"
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
			return errors.New(string(resp.Body()))
		}
		parseTokenResponse(resp.Result().(*authSuccess))
		return nil
	}

	return errors.New("Authentication failed")
}

func (s *restClient) Login(username string, password string) error {
	(*s).username = username
	(*s).password = password
	body := "username=" + username + "&password=" + password + "&client_id=frontend&client_secret=&grant_type=password"
	return auth(body)
}

func (s *restClient) RefreshToken() error {
	body := "client_id=frontend&client_secret=&grant_type=refresh_token&refresh_token=" + s.refreshToken
	return auth(body)
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

func (s *restClient) Create(resourceName string, name string, source string, sourceType string, values map[string]string) (bool, error) {
	return (*s).Apply(resourceName, name, source, sourceType, values, false)
}

func (s *restClient) Update(resourceName string, name string, source string, sourceType string, values map[string]string) (bool, error) {
	return (*s).Apply(resourceName, name, source, sourceType, values, true)
}

func (s *restClient) Apply(resourceName string, name string, source string, sourceType string, values map[string]string, update bool) (bool, error) {

	apply := func() (*resty.Response, error) {
		s.refreshTokenIfNeeded()
		
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
			version = (*s).version
		}

		url, _ := getUrlForResource((*s).url, version, resourceName, "", name, values)
		logging.Info("Requesting url: %v\n", url)
		var resp *resty.Response
		var err error
		if update {
			resp, err = resty.R().
				SetHeader("Content-Type", "application/json").
				SetHeader("Accept", "application/json").
				SetAuthToken((*s).token).
				SetBody(body).
				SetResult(&successResponse{}).
				SetError(&errorResponse{}).
				Put(url)
		} else {
			resp, err = resty.R().
				SetHeader("Content-Type", "application/json").
				SetHeader("Accept", "application/json").
				SetAuthToken((*s).token).
				SetBody(body).
				SetResult(&successResponse{}).
				SetError(&errorResponse{}).
				Post(url)
		}

		return resp, err
	}

	resp, err := fallbackToRefreshToken(apply)

	if err != nil {
		return false, err
	}

	if resp.IsError() {
		return false, getError(resp)
	}
	return true, nil
}

func (s *restClient) fallbackToRefreshToken(f func() (*resty.Response, error)) (*resty.Response, error) {
	resp, err := f()
	if err == nil && resp.IsError() {
		if resp.StatusCode == 401 {
			logging.Info("Got 401, refreshing token...")
			if err := s.RefreshToken(); err != nil {
				log.Fatal("Cannot refresh token - ", err)
			}
			return f()
		} else {
			return resp, getError(resp)
		}
	}
	return resp, err
}

func (s *restClient) Delete(resourceName string, name string, values map[string]string) (bool, error) {
	url, _ := getUrlForResource((*s).url, (*s).version, resourceName, "", name, values)

	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetAuthToken((*s).token).
		SetResult(&successResponse{}).
		SetError(&errorResponse{}).
		Delete(url)

	if err != nil {
		return false, err
	}

	if resp.IsError() {
		return false, getError(resp)
	}
	return true, nil

}

func (s *restClient) GetSpec(resourceName string, name string, outputFormat string, values map[string]string) (string, error) {
	url, _ := getUrlForResource((*s).url, (*s).version, resourceName, "", name, values)

	resp, getResourceError := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetAuthToken((*s).token).
		SetError(&errorResponse{}).
		Get(url)

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

func (s *restClient) Get(resourceName string, name string, outputFormat string, values map[string]string) (string, error) {
	url, _ := getUrlForResource((*s).url, (*s).version, resourceName, "", name, values)

	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetAuthToken((*s).token).
		// SetResult(&successResponse{}). On get Success will be parsed manually
		SetError(&errorResponse{}).
		Get(url)

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

func (s *restClient) List(resourceName string, outputFormat string, values map[string]string, simple bool) (string, error) {
	url, _ := getUrlForResource((*s).url, (*s).version, resourceName, "list", "", values)

	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetAuthToken((*s).token).
		// SetResult(&successResponse{}). On Success list output will be parsed
		SetError(&errorResponse{}).
		Get(url)

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

func (s *restClient) UpdateUserPermission(username string, permission string, values map[string]string) (bool, error) {
	url, _ := getUrlForResource((*s).url, (*s).version, "user-access-permission", "", "", values)
	url += "&user_name=" + username

	permissionBody := models.Permission{
		Read:       strings.Contains(permission, "r"),
		Write:      strings.Contains(permission, "w"),
		Delete:     strings.Contains(permission, "d"),
		EditAccess: strings.Contains(permission, "a"),
	}

	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetAuthToken((*s).token).
		SetBody(permissionBody).
		SetResult(&successResponse{}).
		SetError(&errorResponse{}).
		Post(url)

	if err != nil {
		return false, err
	}

	if resp.IsError() {
		return false, getError(resp)
	}
	return true, nil

}

func (s *restClient) RemovePermissionFromUser(username string, values map[string]string) (bool, error) {
	url, _ := getUrlForResource((*s).url, (*s).version, "user-access-permission", "", "", values)
	url += "&user_name=" + username
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetAuthToken((*s).token).
		SetResult(&authSuccess{}).
		SetError(&errorResponse{}).
		Delete(url)

	if err != nil {
		return false, err
	}

	if resp.IsError() {
		return false, getError(resp)
	}
	return true, nil

}

func (s *restClient) AddRoleToUser(username string, rolename string, values map[string]string) (bool, error) {
	url, _ := getUrlForResource((*s).url, (*s).version, "user-access-role", "", "", values)
	url += "&user_name=" + username + "&role_name=" + rolename
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetAuthToken((*s).token).
		SetResult(&successResponse{}).
		SetError(&errorResponse{}).
		Post(url)

	if err != nil {
		return false, err
	}

	if resp.IsError() {
		return false, getError(resp)
	}
	return true, nil

}

func (s *restClient) RemoveRoleFromUser(username string, rolename string, values map[string]string) (bool, error) {
	url, _ := getUrlForResource((*s).url, (*s).version, "user-access-role", "", "", values)
	url += "&user_name=" + username + "&role_name=" + rolename
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetAuthToken((*s).token).
		SetResult(&authSuccess{}).
		SetError(&errorResponse{}).
		Delete(url)

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
func (s *restClient) Ping() (bool, error) {
	url := (*s).url + "/"
	resty.SetTimeout(5 * time.Second)
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "text/plain, application/json").
		// Should be reachable without a token SetAuthToken((*s).token).
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

func (s *restClient) ReadNotifications(notifications chan<- models.Notification) error {

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u, urlParseError := url.Parse(s.url + "/api/" + s.version + "/notifications?access_token=" + s.token)
	if urlParseError != nil {
		return urlParseError
	}
	u.Scheme = "wss"
	logging.Info("connecting to %s", u.String())
	dialer := websocket.DefaultDialer
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(s.certs))
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
