// Copyright 2019 Kuei-chun Chen. All rights reserved.

package atlas

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/simagix/gox"
)

// GetClusters gets clusters by a group
func (api *API) GetClusters(groupID string) (map[string]interface{}, error) {
	var err error
	var doc map[string]interface{}
	var b []byte

	uri := BaseURL + "/groups/" + groupID + "/clusters"
	if b, err = api.Get(uri); err != nil {
		return nil, err
	}
	json.Unmarshal(b, &doc)
	if api.verbose {
		fmt.Println(gox.Stringify(doc, "", "  "))
	}
	return doc, err
}

// GetCluster gets clusters by a group
func (api *API) GetCluster(groupID string, clusterName string) (map[string]interface{}, error) {
	var err error
	var doc map[string]interface{}
	var b []byte

	uri := BaseURL + "/groups/" + groupID + "/clusters/" + clusterName
	if b, err = api.Get(uri); err != nil {
		return nil, err
	}
	json.Unmarshal(b, &doc)
	if api.verbose {
		fmt.Println(gox.Stringify(doc, "", "  "))
	}
	return doc, err
}

// ClustersDo execute a command
func (api *API) ClustersDo(method string, data string) (string, error) {
	var err error
	var resp *http.Response
	var doc map[string]interface{}
	var b []byte

	if api.groupID == "" {
		return "", errors.New("invalid format ([atlas://]publicKey:privateKey@group)")
	}
	uri := BaseURL + "/groups/" + api.groupID + "/clusters"
	if api.clusterName != "" {
		uri += "/" + api.clusterName
	}
	body := []byte(data)
	headers := map[string]string{}
	headers["Content-Type"] = api.contentType
	headers["Accept"] = api.acceptType
	if resp, err = gox.HTTPDigest(method, uri, api.publicKey, api.privateKey, headers, body); err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err = ioutil.ReadAll(resp.Body)
	json.Unmarshal(b, &doc)
	return gox.Stringify(doc, "", "  "), err
}
