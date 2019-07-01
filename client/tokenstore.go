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
	"io/ioutil"
	"time"

	"github.com/magneticio/vampkubistcli/logging"
	yaml "gopkg.in/yaml.v2"
)

type TokenStore interface {
	Store(token string, timeout int64) error
	Get(token string) (int64, bool)
	Tokens() map[string]int64
	RemoveExpired() error
	Clean() error
}

type FileBackedTokenStore struct {
	Path string
}

func (ts *FileBackedTokenStore) Store(token string, timeout int64) error {
	data, err := ioutil.ReadFile(ts.Path)
	var tokenMap map[string]int64
	if err != nil {
		logging.Info("Token file can not be opened: %v\n", err)
	}
	unmarshalError := yaml.Unmarshal(data, &tokenMap)
	if unmarshalError != nil {
		logging.Info("Token file can not be read: %v\n", unmarshalError)
	}
	if tokenMap == nil {
		tokenMap = make(map[string]int64)
	}
	tokenMap[token] = timeout
	bs, marshalError := yaml.Marshal(tokenMap)
	if marshalError != nil {
		return marshalError
	}
	writeFileError := ioutil.WriteFile(ts.Path, bs, 0644)
	if writeFileError != nil {
		return writeFileError
	}
	return nil
}

func (ts *FileBackedTokenStore) Clean() error {
	tokenMap := make(map[string]int64)
	bs, marshalError := yaml.Marshal(tokenMap)
	if marshalError != nil {
		return marshalError
	}
	writeFileError := ioutil.WriteFile(ts.Path, bs, 0644)
	if writeFileError != nil {
		return writeFileError
	}
	return nil
}

func (ts *FileBackedTokenStore) RemoveExpired() error {
	data, err := ioutil.ReadFile(ts.Path)
	var tokenMap map[string]int64
	if err != nil {
		return err
	}
	unmarshalError := yaml.Unmarshal(data, &tokenMap)
	if unmarshalError != nil {
		return unmarshalError
	}
	if tokenMap == nil {
		return nil
	}
	for token, timeout := range tokenMap {
		if time.Now().Unix() >= timeout {
			delete(tokenMap, token)
		}
	}
	bs, marshalError := yaml.Marshal(tokenMap)
	if marshalError != nil {
		return marshalError
	}
	writeFileError := ioutil.WriteFile(ts.Path, bs, 0644)
	if writeFileError != nil {
		return writeFileError
	}
	return nil
}

func (ts *FileBackedTokenStore) Get(token string) (int64, bool) {
	data, err := ioutil.ReadFile(ts.Path)
	var tokenMap map[string]int64
	if err != nil {
		return 0, false
	}
	unmarshalError := yaml.Unmarshal(data, &tokenMap)
	if unmarshalError != nil {
		return 0, false
	}
	if tokenMap == nil {
		return 0, false
	}
	if timeout, ok := tokenMap[token]; ok {
		return timeout, true
	}
	return 0, false
}

func (ts *FileBackedTokenStore) Tokens() map[string]int64 {
	data, err := ioutil.ReadFile(ts.Path)
	var tokenMap map[string]int64
	if err != nil {
		return make(map[string]int64)
	}
	unmarshalError := yaml.Unmarshal(data, &tokenMap)
	if unmarshalError != nil {
		return make(map[string]int64)
	}
	if tokenMap == nil {
		return make(map[string]int64)
	}
	return tokenMap
}

type InMemoryTokenStore struct {
	tokenMap map[string]int64
}

func (ts *InMemoryTokenStore) Store(token string, timeout int64) error {
	if ts.tokenMap == nil {
		ts.tokenMap = make(map[string]int64)
	}
	ts.tokenMap[token] = timeout
	return nil
}

func (ts *InMemoryTokenStore) Get(token string) (int64, bool) {
	if ts.tokenMap == nil {
		return 0, false
	}
	if timeout, ok := ts.tokenMap[token]; ok {
		return timeout, true
	}
	return 0, false
}

func (ts *InMemoryTokenStore) Tokens() map[string]int64 {
	return ts.tokenMap
}

func (ts *InMemoryTokenStore) RemoveExpired() error {
	if ts.tokenMap == nil {
		return nil
	}
	for token, timeout := range ts.tokenMap {
		if time.Now().Unix() >= timeout {
			delete(ts.tokenMap, token)
		}
	}
	return nil
}

func (ts *InMemoryTokenStore) Clean() error {
	ts.tokenMap = make(map[string]int64)
	return nil
}
