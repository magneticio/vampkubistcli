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
	"errors"
	"fmt"

	"github.com/ghodss/yaml"
	"gopkg.in/resty.v1"
)

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

func NewRestClient(url string, token string) *RestClient {
	return &RestClient{
		url:   url,
		token: token,
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

func getUrlForResource(base string, resourceName string, name string, values map[string]string) (string, error) {
	switch resourceName {
	case "project":
		return base + "/1.0/api/" + "projects" + "?" + resourceName + "_name=" + name, nil
	case "cluster":
		project := values["project"]
		url := base + "/1.0/api/" + "clusters" +
			"?" + "project_name=" + project +
			"&" + resourceName + "_name=" + name
		return url, nil
	case "virtual_cluster":
		project := values["project"]
		cluster := values["cluster"]
		url := base + "/1.0/api/" + "virtual-clusters" +
			"?" + "project_name=" + project +
			"&" + "cluster_name=" + cluster +
			"&" + resourceName + "_name=" + name
		return url, nil
	}
	project := values["project"]
	cluster := values["cluster"]
	virtualCluster := values["virtual_cluster"]
	url := base + "/1.0/api/" + resourceName + "s" +
		"?" + "project_name=" + project +
		"&" + "cluster_name=" + cluster +
		"&" + "virtual_cluster_name=" + virtualCluster +
		"&" + resourceName + "_name=" + name
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
	// resourceType, _ := getResourceType(resourceName)
	// url := (*s).url + "/1.0/api/" + resourceType + "?" + resourceName + "_name=" + name
	url, _ := getUrlForResource((*s).url, resourceName, name, values)
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

func (s *RestClient) Delete(resourceName string, name string) (bool, error) {
	resourceType, _ := getResourceType(resourceName)
	url := (*s).url + "/1.0/api/" + resourceType + "?" + resourceName + "_name=" + name

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

func (s *RestClient) Get(resourceName string, name string) (string, error) {
	resourceType, _ := getResourceType(resourceName)
	url := (*s).url + "/1.0/api/" + resourceType + "?" + resourceName + "_name=" + name

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
		yaml, err_2 := yaml.JSONToYAML(resp.Body())
		if err_2 != nil {
			fmt.Printf("err: %v\n", err_2)
			return "", err_2
		}
		source := string(yaml)
		return source, nil
	} else {
		fmt.Printf("\nError: %v", err)
		return "", err
	}

	return "", nil
}
