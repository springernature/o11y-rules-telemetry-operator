/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"gopkg.in/yaml.v3"
)

type MimirGroupRule map[string]interface{}

type MimirRules struct {
	GroupName string           `yaml:"name"`
	Rules     []MimirGroupRule `yaml:"rules"`
}

type MimirNamespaceRules map[string][]MimirRules

func NewMimirGroupRules(name string, rules []byte) (*MimirRules, error) {
	var rulesJson []MimirGroupRule

	err := json.Unmarshal(rules, &rulesJson)
	groupRules := MimirRules{
		GroupName: name,
		Rules:     rulesJson,
	}
	return &groupRules, err
}

func (mrg *MimirRules) ToYaml() ([]byte, error) {
	return yaml.Marshal(mrg)
}

func (mrg *MimirRules) ToString() (string, error) {
	out, err := mrg.ToYaml()
	return string(out), err
}

type MimirRulerClient struct {
	api       *url.URL
	client    *http.Client
	userAgent string
}

func NewMimirRulerClient(apiUrl *url.URL) *MimirRulerClient {
	//ref: Copy and modify defaults from https://golang.org/src/net/http/transport.go
	//Note: Clients and Transports should only be created once and reused
	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment, // Use proxy setting from env vars
		MaxIdleConns:        100,                       // Maximum total number of idle connections across all hosts
		MaxIdleConnsPerHost: 100,                       // Maximum number of idle connections per host
		Dial: (&net.Dialer{
			// Modify the time to wait for a connection to establish
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		ForceAttemptHTTP2:     false,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 5 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		DisableKeepAlives:     false,
	}
	client := http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}
	ruler := MimirRulerClient{
		api:       apiUrl,
		client:    &client,
		userAgent: "mimirruler-controller/1.0",
	}
	return &ruler
}

func (ruler *MimirRulerClient) makeRequest(request *http.Request) (int, string, error) {
	resp, err := ruler.client.Do(request)
	if err != nil {
		return 0, "error", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, "error", err
	}
	return resp.StatusCode, string(body), err
}

func (ruler *MimirRulerClient) buildRequest(method, path, tenant string, headers map[string]string, payload io.Reader) (*http.Request, error) {
	requestURL, err := ruler.api.Parse(ruler.api.Path + path)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, requestURL.String(), payload)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", ruler.userAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/yaml")
	if tenant != "" {
		req.Header.Set("X-Scope-OrgID", tenant)
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	return req, nil
}

func (ruler *MimirRulerClient) SetGroupRules(tenant, namespace string, rules *MimirRules) (string, error) {
	requestURL := fmt.Sprintf("/config/v1/rules/%s", namespace)
	postBody, _ := rules.ToYaml()
	request, err := ruler.buildRequest(http.MethodPost, requestURL, tenant, nil, bytes.NewReader(postBody))
	if err != nil {
		return "error", err
	}
	statusCode, out, err := ruler.makeRequest(request)
	if statusCode >= 200 && statusCode <= 299 {
		return out, err
	}
	if err == nil {
		return out, fmt.Errorf("HTTP POST error, status code: %d", statusCode)
	}
	return out, fmt.Errorf("HTTP POST, status code %d: %s", statusCode, err.Error())
}

func (ruler *MimirRulerClient) GetGroupRules(tenant, namespace string) ([]MimirRules, error) {
	requestURL := fmt.Sprintf("/config/v1/rules/%s", namespace)
	request, err := ruler.buildRequest(http.MethodGet, requestURL, tenant, nil, nil)
	if err != nil {
		return []MimirRules{}, err
	}
	statusCode, out, err := ruler.makeRequest(request)
	if !(statusCode >= 200 && statusCode <= 299) || (err != nil) {
		if err == nil {
			return []MimirRules{}, fmt.Errorf("HTTP GET error, status code: %d", statusCode)
		}
		return []MimirRules{}, fmt.Errorf("HTTP GET, status code %d: %s", statusCode, err.Error())
	}
	namespaceRules := MimirNamespaceRules{}
	err = yaml.Unmarshal([]byte(out), &namespaceRules)
	if err != nil {
		return []MimirRules{}, err
	}
	// The API returns always empty json even if the namespace is not found
	if rules, found := namespaceRules[namespace]; found {
		return rules, err
	}
	return []MimirRules{}, fmt.Errorf("namespace %s not found", namespace)
}

func (ruler *MimirRulerClient) DeleteGroupRules(tenant, namespace, groupName string) (string, error) {
	requestURL := fmt.Sprintf("/config/v1/rules/%s/%s", namespace, groupName)
	request, err := ruler.buildRequest(http.MethodDelete, requestURL, tenant, nil, nil)
	if err != nil {
		return "error", err
	}
	statusCode, out, err := ruler.makeRequest(request)
	if statusCode >= 200 && statusCode <= 299 {
		return out, err
	}
	if err == nil {
		return out, fmt.Errorf("HTTP DELETE error, status code: %d", statusCode)
	}
	return out, fmt.Errorf("HTTP DELETE, status code %d: %s", statusCode, err.Error())
}

func (ruler *MimirRulerClient) DeleteNamespace(tenant, namespace string) (string, error) {
	requestURL := fmt.Sprintf("/config/v1/rules/%s", namespace)
	request, err := ruler.buildRequest(http.MethodDelete, requestURL, tenant, nil, nil)
	if err != nil {
		return "error", err
	}
	statusCode, out, err := ruler.makeRequest(request)
	if statusCode >= 200 && statusCode <= 299 {
		return out, err
	}
	if err == nil {
		return out, fmt.Errorf("HTTP DELETE, status code: %d", statusCode)
	}
	return out, fmt.Errorf("HTTP DELETE, status code %d: %s", statusCode, err.Error())
}
