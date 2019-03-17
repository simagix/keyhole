// Copyright 2019 Kuei-chun Chen. All rights reserved.

package util

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// HTTPDigest --digest
func HTTPDigest(method string, uri string, username string, password string, headers map[string]string) (*http.Response, error) {
	var err error
	var req *http.Request
	var resp *http.Response

	req, err = http.NewRequest("GET", uri, nil)
	if err != nil {
		return resp, err
	}

	req.SetBasicAuth(username, password)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return resp, err
	}
	defer resp.Body.Close()
	digest := map[string]string{}
	if len(resp.Header["Www-Authenticate"]) > 0 {
		wantedHeaders := []string{"nonce", "realm", "qop"}
		responseHeaders := strings.Split(resp.Header["Www-Authenticate"][0], ",")
		for _, r := range responseHeaders {
			for _, w := range wantedHeaders {
				if strings.Contains(r, w) {
					digest[w] = strings.Split(r, `"`)[1]
				}
			}
		}
	}
	digest["uri"] = uri
	digest["method"] = method
	digest["username"] = username
	digest["password"] = password
	ha1 := hash(digest["username"] + ":" + digest["realm"] + ":" + digest["password"])
	ha2 := hash(digest["method"] + ":" + digest["uri"])
	nonceCount := 00000001
	b := make([]byte, 8)
	io.ReadFull(rand.Reader, b)
	cnonce := fmt.Sprintf("%x", b)[:16]
	response := hash(fmt.Sprintf("%s:%s:%v:%s:%s:%s", ha1, digest["nonce"], nonceCount, cnonce, digest["qop"], ha2))
	authorization := fmt.Sprintf(`Digest username="%s", realm="%s", nonce="%s", uri="%s", cnonce="%s", nc="%v", qop="%s", response="%s"`,
		digest["username"], digest["realm"], digest["nonce"], digest["uri"], cnonce, nonceCount, digest["qop"], response)
	req.Header.Set("Authorization", authorization)
	return http.DefaultClient.Do(req)
}

func hash(s string) string {
	h := md5.New()
	io.WriteString(h, s)
	return hex.EncodeToString(h.Sum(nil))
}
