// Copyright 2019 Kuei-chun Chen. All rights reserved.

package atlas

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/simagix/gox"
)

// AlertsDo execute a command
func (api *API) AlertsDo(method string, data string) (string, error) {
	var err error
	var resp *http.Response
	var doc map[string]interface{}
	var b []byte

	if api.groupID == "" {
		return "", errors.New("invalid format ([atlas://]publicKey:privateKey@group)")
	}
	uri := BaseURL + "/groups/" + api.groupID + "/alertConfigs"
	body := []byte(data)
	headers := map[string]string{}
	headers["Content-Type"] = api.contentType
	headers["Accept"] = api.acceptType
	if resp, err = gox.HTTPDigest(method, uri, api.publicKey, api.privateKey, headers, body); err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err = io.ReadAll(resp.Body)
	json.Unmarshal(b, &doc)
	return gox.Stringify(doc, "", "  "), err
}

// AddAlerts reads from a file and set alerts
func (api *API) AddAlerts(filename string) (string, error) {
	var err error
	var buf []byte
	var str string
	var alerts []map[string]interface{}
	if buf, err = os.ReadFile(filename); err != nil {
		return "", err
	}

	json.Unmarshal(buf, &alerts)
	for _, alert := range alerts {
		str, err = api.AlertsDo("POST", gox.Stringify(alert))
		fmt.Println(str)
	}
	return "", err
}
