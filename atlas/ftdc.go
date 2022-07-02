// Copyright 2019 Kuei-chun Chen. All rights reserved.

package atlas

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/simagix/gox"
)

// FTDC stores Atlas logs API info
type FTDC struct {
}

// DownloadFTDC downloads logs
func (api *API) DownloadFTDC() (string, error) {
	var err error
	var resp *http.Response
	doc := map[string]interface{}{}

	downloadSize := 10 * 1024 * 1024
	clusterType := ""
	for _, param := range api.params {
		s := strings.Split(param, "=")
		if s[0] == "size" {
			downloadSize, _ = strconv.Atoi(s[1])
		} else if s[0] == "clusterType" {
			clusterType = s[1]
		}
	}
	if clusterType == "" {
		tokens := strings.Split(api.clusterName, "-")
		if len(tokens) > 4 && tokens[len(tokens)-2] == "node" {
			clusterType = "PROCESS"
		} else if len(tokens) > 2 && tokens[len(tokens)-2] == "shard" {
			clusterType = "REPLICASET"
		} else {
			clusterType = "CLUSTER"
		}
	}
	doc["resourceType"] = clusterType
	doc["resourceName"] = api.clusterName
	doc["redacted"] = true
	doc["sizeRequestedPerFileBytes"] = downloadSize
	doc["logTypes"] = []string{"FTDC"}
	body, _ := json.Marshal(doc)
	uri := BaseURL + "/groups/" + api.groupID + "/logCollectionJobs"
	headers := map[string]string{}
	headers["Content-Type"] = api.contentType
	headers["Accept"] = api.acceptType
	if resp, err = gox.HTTPDigest("POST", uri, api.publicKey, api.privateKey, headers, body); err != nil {
		return "", err
	}
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	json.Unmarshal(body, &doc)
	if doc["id"] == nil {
		return "", errors.New("error creating logCollection job")
	}
	jobID := doc["id"].(string)
	fmt.Println(gox.Stringify(doc, "", "  "))

	// Get logs ready
	status := "IN_PROGRESS"
	for status == "IN_PROGRESS" {
		fmt.Println(status)
		time.Sleep(10 * time.Second)
		uri = BaseURL + "/groups/" + api.groupID + "/logCollectionJobs/" + jobID + "?verbose=true&pretty=true"
		body, _ = api.Get(uri)
		json.Unmarshal(body, &doc)
		status = doc["status"].(string)
	}

	// download to diagnostic.tar.gz
	api.SetAcceptType(ApplicationGZip)
	uri = BaseURL + "/groups/" + api.groupID + "/logCollectionJobs/" + jobID + "/download"
	body, _ = api.Get(uri)
	fname := api.clusterName + "-diagnostic.tar.gz"

	ioutil.WriteFile(fname, body, 0644)

	// delete the log collection job
	uri = BaseURL + "/groups/" + api.groupID + "/logCollectionJobs/" + jobID
	headers["Content-Type"] = api.contentType
	headers["Accept"] = api.acceptType
	if resp, err = gox.HTTPDigest("DELETE", uri, api.publicKey, api.privateKey, headers, body); err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ = ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &doc)
	fmt.Println(gox.Stringify(doc, "", "  "))
	return fmt.Sprintf("FTDC archive was written to %v", fname), err
}
