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
	"fmt"

	"gopkg.in/resty.v1"
)

type RestClient struct {
	server   string
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

func NewRestClient(server string, token string) *RestClient {
	return &RestClient{
		server: server,
		token:  token,
	}
}

func (s *RestClient) Login(username string, password string) (string, error) {
	(*s).username = username
	(*s).password = password
	url := "http://" + s.server + "/oauth/access_token"
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
