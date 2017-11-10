/*
Copyright 2016 Intel Corporation

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

package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	log "github.com/sirupsen/logrus"
)

// Define a client for collecting metrics from a Mesos master/agent over HTTP.
type Client struct {
	httpClient *http.Client
	host       string
	path       string
}

// Return a new instance of Client.
func NewClient(host string, path string, timeout time.Duration) *Client {
	log.Debug("Creating a new instance of the Mesos plugin HTTP client")
	return &Client{
		httpClient: &http.Client{Timeout: timeout},
		host:       host,
		path:       path,
	}
}

// Return the URL for this client as a string. Note that this isn't specifically required for this client, but might
// be useful if you want to retrieve the actual URL for logging, etc throughout this plugin.
func (c *Client) URL() string {
	u := url.URL{Scheme: "http", Host: c.host, Path: c.path}
	return u.String()
}

// Fetch JSON from the API endpoint, unmarshal it, and return it to the provided 'target'.
func (c *Client) Fetch(target interface{}) error {
	log.Debug("Fetching data from ", c.URL())
	resp, err := http.Get(c.URL())
	if err != nil {
		log.Error(err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		e := fmt.Errorf("fetch error: %s", resp.Status)
		log.Error(e)
		return e
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		e := fmt.Errorf("read error: %s: %v\n", c.URL(), err)
		log.Error(e)
		return e
	}

	if err := json.Unmarshal(b, &target); err != nil {
		e := fmt.Errorf("unmarshal error: %s: %v\n", b, err)
		log.Error(e)
		return e
	}

	return nil
}
