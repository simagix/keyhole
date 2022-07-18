// Copyright 2019 Kuei-chun Chen. All rights reserved.

package atlas

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/simagix/gox"
)

// Project stores project info
type Project struct {
	ID    string
	Name  string
	OrgID string
}

// GetGroups get processes of a user
func (api *API) GetGroups() (map[string]interface{}, error) {
	var err error
	var doc map[string]interface{}
	var b []byte

	uri := BaseURL + "/groups"
	if b, err = api.Get(uri); err != nil {
		return doc, err
	}
	json.Unmarshal(b, &doc)
	if api.verbose {
		fmt.Println(gox.Stringify(doc, "", "  "))
	}
	return doc, err
}

// GetGroupsByID get processes of a user
func (api *API) GetGroupsByID(groupID string) (map[string]interface{}, error) {
	var err error
	var doc map[string]interface{}
	var b []byte

	uri := BaseURL + "/groups/" + groupID
	if b, err = api.Get(uri); err != nil {
		return nil, err
	}
	json.Unmarshal(b, &doc)
	return doc, err
}

// GetProjects returns an array of group IDs
func (api *API) GetProjects() ([]Project, error) {
	var projects []Project
	var err error
	var doc map[string]interface{}
	if doc, err = api.GetGroups(); err != nil {
		return projects, err
	}
	_, ok := doc["results"]
	if !ok {
		return projects, errors.New(gox.Stringify(doc))
	}
	results := doc["results"].([]interface{})
	for _, result := range results {
		var project Project
		b, _ := json.Marshal(result)
		json.Unmarshal(b, &project)
		projects = append(projects, project)
	}
	return projects, err
}
