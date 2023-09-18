// Copyright 2020-present Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// GenerateMaobiReport outputs an HTML from Maobi
func GenerateMaobiReport(maobiURL string, data []byte, ofile string) error {
	var err error
	var murl *url.URL
	if maobiURL == "" || ofile == "" {
		return nil
	}
	if murl, err = url.Parse(maobiURL); err != nil {
		return err
	}
	os.Mkdir(htmldir, 0755)
	i := strings.Index(ofile, ".bson.gz")
	filename := strings.Replace(ofile[:i]+".html", outdir, htmldir, 1)
	dial := net.Dialer{Timeout: 2 * time.Second}
	var c net.Conn
	to := murl.Hostname()
	if murl.Port() != "" {
		to += ":" + murl.Port()
	}
	if c, err = dial.Dial("tcp", to); err != nil {
		return err
	}
	c.Close()
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	fileWriter, err := bodyWriter.CreateFormFile("file", ofile)
	if err != nil {
		return err
	}
	fileWriter.Write(data)
	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()
	resp, err := http.Post(maobiURL, contentType, bodyBuf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	os.WriteFile(filename, body, 0644)
	fmt.Printf("HTML report written to %v\n", filename)
	return err
}
