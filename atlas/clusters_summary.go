// Copyright 2019 Kuei-chun Chen. All rights reserved.

package atlas

import (
	"errors"
	"fmt"
	"strings"

	"github.com/simagix/gox"
)

// GetClustersSummary retrieve clusters summary for a user
func (api *API) GetClustersSummary() (string, error) {
	var err error
	var buffers []string
	var projects []Project
	var doc map[string]interface{}

	if projects, err = api.GetProjects(); err != nil {
		return "", err
	}

	for _, project := range projects {
		if doc, err = api.GetClusters(project.ID); err != nil {
			return "", err
		}
		if _, ok := doc["results"]; !ok {
			return "", errors.New(gox.Stringify(doc))
		}
		clusters := doc["results"]
		buffers = append(buffers, fmt.Sprint("- Project: ", project.Name))
		buffers = append(buffers, fmt.Sprint("- Group ID: ", project.ID))
		for _, cluster := range clusters.([]interface{}) {
			m := cluster.(map[string]interface{})
			buffers = append(buffers, fmt.Sprint("  - cluster name: ", m["name"]))
			buffers = append(buffers, fmt.Sprint("    - ", m["mongoDBVersion"], ", ", m["clusterType"], ", ", m["srvAddress"]))
		}
		if doc, err = api.GetProcesses(project.ID); err != nil {
			return strings.Join(buffers, "\n"), err
		}
		if _, ok := doc["results"]; !ok {
			return strings.Join(buffers, "\n"), err
		}
		processes := doc["results"]
		buffers = append(buffers, "\n  - Processes:")
		replicaSetName := ""
		for _, process := range processes.([]interface{}) {
			maps := process.(map[string]interface{})
			if maps["typeName"].(string) == "NO_DATA" {
				continue
			}
			if maps["replicaSetName"] != nil && replicaSetName != maps["replicaSetName"].(string) {
				replicaSetName = maps["replicaSetName"].(string)
				buffers = append(buffers, fmt.Sprint("    - Replica set name: ", replicaSetName))
			}
			buffers = append(buffers, fmt.Sprint("      - ", maps["hostname"], " (", maps["typeName"], ")"))
		}
		buffers = append(buffers, "\n")
	}
	return strings.Join(buffers, "\n"), err
}
