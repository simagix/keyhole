// Copyright 2019 Kuei-chun Chen. All rights reserved.

package atlas

import (
	"encoding/json"
	"fmt"

	"github.com/simagix/gox"
)

// GetProcesses get processes of a user
func (api *API) GetProcesses(groupID string) (map[string]interface{}, error) {
	var err error
	var doc map[string]interface{}
	var b []byte

	uri := BaseURL + "/groups/" + groupID + "/processes"
	if b, err = api.Get(uri); err != nil {
		return nil, err
	}
	json.Unmarshal(b, &doc)
	if api.verbose {
		fmt.Println(gox.Stringify(doc, "", "  "))
	}
	return doc, err
}
