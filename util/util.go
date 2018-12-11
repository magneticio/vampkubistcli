package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/ghodss/yaml"
)

/*
This function allows using a filepath or http/s url to get resource from
*/
func UseSourceUrl(resourceUrl string) (string, error) {
	u, err := url.ParseRequestURI(resourceUrl)
	if err != nil {
		file, err := ioutil.ReadFile(resourceUrl) // just pass the file name
		if err != nil {
			return "", err
		}
		source := string(file)
		return source, nil
	}
	scheme := strings.ToLower(u.Scheme)
	fmt.Println("scheme: " + scheme)
	if scheme == "http" || scheme == "https" {
		resp, err := http.Get(resourceUrl)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		source := string(contents)
		fmt.Println(source)
		return source, nil
	}
	return "", errors.New("Not Supported protocol :" + scheme)
}

func Convert(inputFormat string, outputFormat string, input string) (string, error) {
	if inputFormat == outputFormat {
		return input, nil
	}

	inputSource := []byte(input)
	if inputFormat == "yaml" {
		json, err := yaml.YAMLToJSON(inputSource)
		if err != nil {
			fmt.Printf("err: %v\n", err)
			return "", err
		}
		inputSource = json
	}

	// convert everything to json as byte

	outputSourceString := ""
	if outputFormat == "yaml" {
		yaml, errYaml := yaml.JSONToYAML(inputSource)
		if errYaml != nil {
			fmt.Printf("YAML conversion error: %v\n", errYaml)
			return "", errYaml
		}
		outputSourceString = string(yaml)
	} else {
		var prettyJSON bytes.Buffer
		indentError := json.Indent(&prettyJSON, inputSource, "", "    ")
		if indentError != nil {
			log.Println("JSON parse error: ", indentError)
			return "", indentError
		}
		outputSourceString = string(prettyJSON.Bytes())
	}
	return outputSourceString, nil
}
