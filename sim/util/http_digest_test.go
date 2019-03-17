// Copyright 2019 Kuei-chun Chen. All rights reserved.

package util

import (
	"os"
	"strings"
	"testing"
)

const BaseURL = "https://cloud.mongodb.com/api/atlas/v1.0"

func TestHTTPDigest(t *testing.T) {
	login := os.Getenv("ATLAS_AUTH")
	idx := strings.Index(login, ":")
	headers := map[string]string{}
	headers["Content-Type"] = "application/json"
	headers["Accept"] = "application/json"
	resp, err := HTTPDigest("GET", BaseURL, login[:idx], login[idx+1:], headers)
	defer resp.Body.Close()
	if resp.StatusCode == 401 {
		t.Fatal(resp.StatusCode)
	} else if err != nil {
		t.Fatal(err)
	}
}
