// Copyright 2019 Kuei-chun Chen. All rights reserved.

package atlas

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/simagix/keyhole/mdb"
	"github.com/simagix/keyhole/sim/util"
)

// Summary info
type Summary struct {
	apiKey   string
	userName string
	verbose  bool
}

// NewSummary establish Atlas credential
func NewSummary(user string) *Summary {
	i := strings.LastIndex(user, ":")
	return &Summary{userName: user[:i], apiKey: user[i+1:]}
}

// SetVerbose -
func (su *Summary) SetVerbose(verbose bool) {
	su.verbose = verbose
}

// GetSummary returns Atlas info
func (su *Summary) GetSummary() (string, error) {
	var err error
	var roles []interface{}
	var buffers []string

	if roles, err = su.getRolesByName(); err != nil {
		return "", err
	}

	for _, role := range roles {
		groupID := role.(map[string]interface{})["groupId"]
		if groupID != nil {
			var clusters []interface{}
			if clusters, err = su.getClusters(groupID.(string)); err != nil {
				return "", err
			}
			var processes []interface{}
			if processes, err = su.getProcesses(groupID.(string)); err != nil {
				return "", err
			}
			buffers = append(buffers, fmt.Sprint("- Group: ", groupID))
			for _, cluster := range clusters {
				m := cluster.(map[string]interface{})
				buffers = append(buffers, fmt.Sprint("  - cluster name: ", m["name"]))
				buffers = append(buffers, fmt.Sprint("    - ", m["mongoDBVersion"], ", ", m["clusterType"], ", ", m["srvAddress"]))
				buffers = append(buffers, fmt.Sprint("    - Hosts:"))
				for _, process := range processes {
					maps := process.(map[string]interface{})
					if strings.Index(strings.ToLower(maps["hostname"].(string)), strings.ToLower(m["name"].(string)+"-")) == 0 {
						buffers = append(buffers, fmt.Sprint("      - ", maps["hostname"], " (", maps["typeName"], ")"))
					}
				}
			}
			buffers = append(buffers, "\n")
		}
	}

	// downloads := `curl --user $ATLAS_AUTH --digest --header 'Accept: application/gzip' --output "${hostname}.gz" \
	//   --request GET "https://cloud.mongodb.com/api/atlas/v1.0/groups/${group_id}/clusters/${hostname}/logs/mongodb.gz"`
	downloads := `keyhole --loginfo atlas://{user_name}:{api_key}@{group_id}/{cluster_name}"`
	buffers = append(buffers, fmt.Sprint("Usage: ", downloads))
	return strings.Join(buffers, "\n"), err
}

func (su *Summary) getRolesByName() ([]interface{}, error) {
	var err error
	var resp *http.Response
	var doc map[string]interface{}
	var b []byte
	var roles []interface{}

	uri := BaseURL + "/users/byName/" + su.userName
	headers := map[string]string{}
	headers["Content-Type"] = ApplicationJSON
	headers["Accept"] = "application/json"
	if resp, err = util.HTTPDigest("GET", uri, su.userName, su.apiKey, headers); err != nil {
		return roles, err
	}
	defer resp.Body.Close()
	if b, err = ioutil.ReadAll(resp.Body); err != nil {
		return roles, err
	}
	json.Unmarshal(b, &doc)
	if su.verbose == true {
		fmt.Println(mdb.Stringify(doc, "", "  "))
	}
	roles = doc["roles"].([]interface{})
	return roles, err
}

func (su *Summary) getClusters(groupID string) ([]interface{}, error) {
	var err error
	var resp *http.Response
	var doc map[string]interface{}
	var b []byte
	var results []interface{}

	uri := BaseURL + "/groups/" + groupID + "/clusters"
	headers := map[string]string{}
	headers["Content-Type"] = ApplicationJSON
	headers["Accept"] = ApplicationJSON
	if resp, err = util.HTTPDigest("GET", uri, su.userName, su.apiKey, headers); err != nil {
		return results, err
	}
	defer resp.Body.Close()
	if b, err = ioutil.ReadAll(resp.Body); err != nil {
		return results, err
	}
	json.Unmarshal(b, &doc)
	if su.verbose == true {
		fmt.Println(mdb.Stringify(doc, "", "  "))
	}
	results = doc["results"].([]interface{})
	return results, err
}

func (su *Summary) getProcesses(groupID string) ([]interface{}, error) {
	var err error
	var resp *http.Response
	var doc map[string]interface{}
	var b []byte
	var results []interface{}

	uri := BaseURL + "/groups/" + groupID + "/processes"
	headers := map[string]string{}
	headers["Content-Type"] = ApplicationJSON
	headers["Accept"] = ApplicationJSON
	if resp, err = util.HTTPDigest("GET", uri, su.userName, su.apiKey, headers); err != nil {
		return results, err
	}
	defer resp.Body.Close()
	if b, err = ioutil.ReadAll(resp.Body); err != nil {
		return results, err
	}
	json.Unmarshal(b, &doc)
	if su.verbose == true {
		fmt.Println(mdb.Stringify(doc, "", "  "))
	}
	results = doc["results"].([]interface{})
	return results, err
}
