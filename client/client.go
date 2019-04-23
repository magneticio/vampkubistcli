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
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ghodss/yaml"
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
	url      string
	version  string
	username string
	password string
	token    string
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

func NewRestClient(url string, token string, version string, isVerbose bool, cert string) *restClient {
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
	return &restClient{
		url:     url,
		token:   token,
		version: version,
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

func (s *restClient) Login(username string, password string) (string, error) {
	(*s).username = username
	(*s).password = password
	url := (*s).url + "/oauth/access_token"
	// fmt.Printf("user login with username: %v password: %v\n", username, password)
	body := "username=" + username + "&password=" + password + "&client_id=frontend&client_secret=&grant_type=password"
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
		(*s).token = resp.Result().(*authSuccess).AccessToken
		return (*s).token, nil
	}

	return "", errors.New("Token retrievel failed")

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
		queryParams = queryParams + "&" + key + "_name=" + value
	}

	timestamp := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	url := base + "/api/" + version + "/" + strings.Replace(resourceName, "_", "-", -1) + "s" + subPath +
		"?" + "t=" + timestamp +
		queryParams +
		namedParameter
	return url, nil
}

func (s *restClient) Create(resourceName string, name string, source string, sourceType string, values map[string]string) (bool, error) {
	return (*s).Apply(resourceName, name, source, sourceType, values, false)
}

func (s *restClient) Update(resourceName string, name string, source string, sourceType string, values map[string]string) (bool, error) {
	return (*s).Apply(resourceName, name, source, sourceType, values, true)
}

func (s *restClient) Apply(resourceName string, name string, source string, sourceType string, values map[string]string, update bool) (bool, error) {

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

	if err != nil {
		return false, err
	}

	if resp.IsError() {
		return false, getError(resp)
	}
	return true, nil

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
