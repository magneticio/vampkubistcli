package util

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/ghodss/yaml"
	"github.com/yalp/jsonpath"
	"golang.org/x/crypto/ssh/terminal"
)

/*
This function allows using a filepath or http/s url to get resource from
*/
func UseSourceUrl(resourceUrl string) (string, error) {
	u, err := url.ParseRequestURI(resourceUrl)
	if err != nil || u.Scheme == "" {
		file, err := ioutil.ReadFile(resourceUrl) // just pass the file name
		if err != nil {
			return "", err
		}
		source := string(file)
		return source, nil
	}
	scheme := strings.ToLower(u.Scheme)
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
			return "", err
		}
		inputSource = json
	}

	// convert everything to json as byte

	outputSourceString := ""
	if outputFormat == "yaml" {
		yaml, errYaml := yaml.JSONToYAML(inputSource)
		if errYaml != nil {
			// fmt.Printf("YAML conversion error: %v\n", errYaml)
			return "", errYaml
		}
		outputSourceString = string(yaml)
	} else {
		var prettyJSON bytes.Buffer
		indentError := json.Indent(&prettyJSON, inputSource, "", "    ")
		if indentError != nil {
			// log.Println("JSON parse error: ", indentError)
			return "", indentError
		}
		outputSourceString = string(prettyJSON.Bytes())
	}
	return outputSourceString, nil
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath string, url string) error {

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func GetJsonPath(source string, sourceFormat string, jsonPath string) (string, error) {
	var jsonInterface map[string]interface{}
	err := json.Unmarshal([]byte(source), &jsonInterface)
	if err != nil {
		return "", err
	}
	resultPath, err := jsonpath.Read(jsonInterface, jsonPath)
	if err != nil {
		return "", err
	}
	str, ok := resultPath.(string)
	if !ok {
		return "", errors.New("There is no string representation for " + jsonPath)
	}
	return str, nil
}

func VerifyCertForHost(resourceUrl string, cert string) error {
	u, err_url := url.ParseRequestURI(resourceUrl)
	if err_url != nil {
		return err_url
	}
	host, _, _ := net.SplitHostPort(u.Host)
	block, _ := pem.Decode([]byte(cert))
	if block == nil {
		return errors.New("failed to decode certificate")
	}
	crt, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return err
	}
	opts := x509.VerifyOptions{DNSName: host, Roots: x509.NewCertPool()}
	opts.Roots.AddCert(crt)
	_, err = crt.Verify(opts)
	return err
}

func GetParameterFromTerminalAsSecret(text1 string, text2 string, errorText string) (string, error) {
	fmt.Println(text1)
	byteInput1, errInput1 := terminal.ReadPassword(int(syscall.Stdin))
	if errInput1 != nil {
		return "", errInput1
	}
	fmt.Println()
	input1 := string(byteInput1)
	fmt.Println(text2)
	byteInput2, errInput2 := terminal.ReadPassword(int(syscall.Stdin))
	if errInput2 != nil {
		return "", errInput2
	}
	fmt.Println()
	input2 := string(byteInput2)
	if input1 != input2 {
		return "", errors.New(errorText)
	}
	return input1, nil
}

var src = rand.NewSource(time.Now().UnixNano())

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func RandomString(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}
