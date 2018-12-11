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
	"fmt"
	"log"
	"strings"

	"github.com/ghodss/yaml"
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

type RestClient struct {
	url      string
	username string
	password string
	token    string
}

type AuthSuccess struct {
	/* variables */
	TokenType    string `json:"token_type"`
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

type AuthError struct {
	/* variables */
}

type Named struct {
	Name string `json:"name"`
}

type Metadata struct {
	Metadata map[string]string `json:"metadata"`
}

type VampService struct {
	Gateways []string `json:"gateways"`
	Hosts    []string `json:"hosts"`
	Routes   []Route  `json:"routes"`
}

type Route struct {
	Protocol string   `json:"protocol"`
	Weights  []Weight `json:"weights"`
}

type Weight struct {
	Destination string `json:"destination"`
	Port        int64  `json:"port"`
	Version     string `json:"version"`
	Weight      int64  `json:"weight"`
}

type CanaryRelease struct {
	VampService  string            `json:"vampService"`
	Destination  string            `json:"destination"`
	Subset       string            `json:"subset"`
	SubsetLabels map[string]string `json:"subsetLabels"`
}

func NewRestClient(url string, token string, isDebug bool) *RestClient {
	resty.SetDebug(isDebug)
	return &RestClient{
		url:   url,
		token: token,
	}
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
	} else {
		return resourceString
	}
}

func (s *RestClient) Login(username string, password string) (string, error) {
	(*s).username = username
	(*s).password = password
	url := (*s).url + "/oauth/access_token"
	// fmt.Printf("user login with username: %v password: %v\n", username, password)
	body := "username=" + username + "&password=" + password + "&client_id=frontend&client_secret=&grant_type=password"
	resp, err := resty.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded; charset=utf-8").
		SetHeader("Accept", "application/json").
		SetBody([]byte(body)).
		SetResult(&AuthSuccess{}). // or SetResult(AuthSuccess{}).
		SetError(&AuthError{}).    // or SetError(AuthError{}).
		Post(url)

	if err == nil {
		// fmt.Printf("\nAccess Token: %v", resp.Result().(*AuthSuccess).AccessToken)
		(*s).token = resp.Result().(*AuthSuccess).AccessToken
	} else {
		fmt.Printf("\nError: %v", err)
		return "", err
	}
	// explore response object
	/*
		fmt.Printf("\nError: %v", err)
		fmt.Printf("\nResponse Status Code: %v", resp.StatusCode())
		fmt.Printf("\nResponse Status: %v", resp.Status())
		fmt.Printf("\nResponse Time: %v", resp.Time())
		fmt.Printf("\nResponse Received At: %v", resp.ReceivedAt())
		fmt.Printf("\nResponse Body: %v", resp) // or resp.String() or string(resp.Body())
		fmt.Printf("\n")
	*/

	return (*s).token, nil
}

func getResourceType(resourceName string) (string, error) {
	if resourceName == "project" {
		return "projects", nil
	}
	return "", errors.New("no resource Type")
}

func getUrlForResource(base string, resourceName string, subCommand string, name string, values map[string]string) (string, error) {
	resourceName = ResourceTypeConversion(resourceName)
	subPath := ""
	namedParameter := ""
	if subCommand != "" {
		subPath = "/" + subCommand
	}
	if name != "" {
		namedParameter = "&" + resourceName + "_name=" + name
	}
	applicationParameter := ""
	application := values["application"]
	if application != "" {
		applicationParameter = "&" + "application_name=" + application
	}
	switch resourceName {
	case "project":
		return base + "/1.0/api/" + "projects" + subPath + "?time=-1" + namedParameter, nil
	case "cluster":
		project := values["project"]
		url := base + "/1.0/api/" + "clusters" + subPath +
			"?" + "project_name=" + project +
			namedParameter
		return url, nil
	case "virtual_cluster":
		project := values["project"]
		cluster := values["cluster"]
		url := base + "/1.0/api/" + "virtual-clusters" + subPath +
			"?" + "project_name=" + project +
			"&" + "cluster_name=" + cluster +
			namedParameter
		return url, nil
	case "virtual_service":
		project := values["project"]
		cluster := values["cluster"]
		url := base + "/1.0/api/" + "virtual-services" + subPath +
			"?" + "project_name=" + project +
			"&" + "cluster_name=" + cluster +
			namedParameter
		return url, nil
	}
	project := values["project"]
	cluster := values["cluster"]
	virtualCluster := values["virtual_cluster"]
	url := base + "/1.0/api/" + strings.Replace(resourceName, "_", "-", -1) + "s" + subPath +
		"?" + "project_name=" + project +
		"&" + "cluster_name=" + cluster +
		"&" + "virtual_cluster_name=" + virtualCluster +
		applicationParameter +
		namedParameter
	return url, nil
	// return "", errors.New("no resource Type")
}

func (s *RestClient) Create(resourceName string, name string, source string, sourceType string, values map[string]string) (bool, error) {
	return (*s).Apply(resourceName, name, source, sourceType, values, false)
}

func (s *RestClient) Update(resourceName string, name string, source string, sourceType string, values map[string]string) (bool, error) {
	return (*s).Apply(resourceName, name, source, sourceType, values, true)
}

func (s *RestClient) Apply(resourceName string, name string, source string, sourceType string, values map[string]string, update bool) (bool, error) {
	url, _ := getUrlForResource((*s).url, resourceName, "", name, values)
	// fmt.Printf("url: %v\n", url)

	if sourceType == "yaml" {
		json, err := yaml.YAMLToJSON([]byte(source))
		if err != nil {
			fmt.Printf("err: %v\n", err)
			return false, err
		}
		source = string(json)
	}

	body := source

	var resp *resty.Response
	var err error
	if update {
		resp, err = resty.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Accept", "application/json").
			SetAuthToken((*s).token).
			SetBody([]byte(body)).
			Put(url)
	} else {
		resp, err = resty.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Accept", "application/json").
			SetAuthToken((*s).token).
			SetBody([]byte(body)).
			Post(url)
	}

	if err == nil {
		fmt.Printf("\nResult: %v\n", resp)
		return true, nil
	} else {
		fmt.Printf("\nError: %v", err)
		return false, err
	}

	return false, nil
}

func (s *RestClient) Delete(resourceName string, name string, values map[string]string) (bool, error) {
	url, _ := getUrlForResource((*s).url, resourceName, "", name, values)

	// body := source
	resp, err := resty.R().
		// SetHeader("Content-Type", "application/x-www-form-urlencoded; charset=utf-8").
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetAuthToken((*s).token).
		// SetBody([]byte(body)).
		// SetResult(&AuthSuccess{}). // or SetResult(AuthSuccess{}).
		// SetError(&AuthError{}).    // or SetError(AuthError{}).
		Delete(url)

	if err == nil {
		fmt.Printf("\nResult: %v\n", resp)
		return true, nil
	} else {
		fmt.Printf("\nError: %v", err)
		return false, err
	}

	return false, nil
}

func (s *RestClient) Get(resourceName string, name string, outputFormat string, values map[string]string) (string, error) {
	url, _ := getUrlForResource((*s).url, resourceName, "", name, values)

	resp, err := resty.R().
		// SetHeader("Content-Type", "application/x-www-form-urlencoded; charset=utf-8").
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetAuthToken((*s).token).
		// SetBody([]byte(body)).
		// SetResult(&AuthSuccess{}). // or SetResult(AuthSuccess{}).
		// SetError(&AuthError{}).    // or SetError(AuthError{}).
		Get(url)

	if err == nil {
		// fmt.Printf("\nResult: %v\n", resp)
		source := ""
		if outputFormat == "yaml" {
			yaml, err_2 := yaml.JSONToYAML(resp.Body())
			if err_2 != nil {
				fmt.Printf("err: %v\n", err_2)
				return "", err_2
			}
			source = string(yaml)
		} else {
			var prettyJSON bytes.Buffer
			error := json.Indent(&prettyJSON, resp.Body(), "", "    ")
			if error != nil {
				log.Println("JSON parse error: ", error)
				return "", error
			}
			source = string(prettyJSON.Bytes())
		}
		return source, nil
	} else {
		fmt.Printf("\nError: %v", err)
		return "", err
	}

	return "", nil
}

func (s *RestClient) List(resourceName string, outputFormat string, values map[string]string, simple bool) (string, error) {
	url, _ := getUrlForResource((*s).url, resourceName, "list", "", values)

	resp, err := resty.R().
		// SetHeader("Content-Type", "application/x-www-form-urlencoded; charset=utf-8").
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetAuthToken((*s).token).
		// SetBody([]byte(body)).
		// SetResult(&AuthSuccess{}). // or SetResult(AuthSuccess{}).
		// SetError(&AuthError{}).    // or SetError(AuthError{}).
		Get(url)

	if err == nil {
		// fmt.Printf("\nResult: %v\n", resp)
		responseBody := resp.Body()
		if simple {
			var r []Named
			err := json.Unmarshal([]byte(responseBody), &r)
			if err != nil {
				fmt.Printf("Error: %v\n", string(responseBody))
				return "", errors.New(string(responseBody))
			}
			responseBody, err = json.Marshal(r)
			if err != nil {
				fmt.Printf("Error: %v", err)
				return "", err
			}
		}

		source := ""
		if outputFormat == "yaml" {
			yaml, err_2 := yaml.JSONToYAML(responseBody)
			if err_2 != nil {
				fmt.Printf("err: %v\n", err_2)
				return "", err_2
			}
			source = string(yaml)
		} else {
			var prettyJSON bytes.Buffer
			error := json.Indent(&prettyJSON, responseBody, "", "    ")
			if error != nil {
				log.Println("JSON parse error: ", error)
				return "", error
			}
			source = string(prettyJSON.Bytes())
		}
		return source, nil
	} else {
		fmt.Printf("\nError: %v", err)
		return "", err
	}

	return "", nil
}
