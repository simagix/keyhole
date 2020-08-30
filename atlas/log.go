// Copyright 2019 Kuei-chun Chen. All rights reserved.

package atlas

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/simagix/gox"
)

type entity struct {
	hostname string
	log      string
}

// DownloadLogs downloads logs
func (api *API) DownloadLogs() ([]string, error) {
	var err error
	var filenames []string
	var doc map[string]interface{}

	if doc, err = api.GetProcesses(api.groupID); err != nil {
		return filenames, err
	}
	_, ok := doc["results"]
	if !ok {
		return filenames, errors.New(gox.Stringify(doc))
	}
	processes := doc["results"]
	startDate := time.Time{}
	endDate := time.Time{}
	hostname := ""
	for _, param := range api.params {
		s := strings.Split(param, "=")
		if s[0] == "startDate" {
			startDate, _ = time.Parse("2006-01-02", s[1])
		} else if s[0] == "endDate" {
			endDate, _ = time.Parse("2006-01-02", s[1])
			endDate = endDate.Add(time.Hour * 24)
		} else if s[0] == "hostname" {
			hostname = s[1]
		}
	}

	if endDate.IsZero() && startDate.IsZero() {
		endDate = time.Now()
		startDate = endDate.Add(time.Hour * -24)
	} else if endDate.IsZero() {
		endDate = startDate.Add(time.Hour * 24)
	} else if startDate.IsZero() {
		startDate = endDate.Add(time.Hour * -24)
	}
	log.Println("download files from", startDate.Format("2006.01.02 15:04:05"), "to", endDate.Format("2006.01.02 15:04:05"))
	hosts := []entity{}
	for _, process := range processes.([]interface{}) {
		maps := process.(map[string]interface{})
		process := maps["typeName"].(string)
		host := maps["hostname"].(string)
		if hostname != "" && hostname != host {
			continue
		}
		if process == "REPLICA_PRIMARY" || process == "REPLICA_SECONDARY" {
			hosts = append(hosts, entity{hostname: host, log: "mongodb.gz"})
		} else if process == "SHARD_MONGOS" {
			hosts = append(hosts, entity{hostname: host, log: "mongos.gz"})
		}
	}

	for _, host := range hosts {
		filename := host.hostname + "-" + host.log
		uri := BaseURL + "/groups/" + api.groupID + "/clusters/" + host.hostname + "/logs/" + host.log
		uri += "?startDate=" + fmt.Sprintf("%v", startDate.Unix()) + "&endDate=" + fmt.Sprintf("%v", endDate.Unix())
		if api.verbose {
			log.Println("download from", uri)
		}
		var b []byte
		api.SetAcceptType(ApplicationGZip)
		if b, err = api.Get(uri); err != nil {
			log.Println(err)
			continue
		}
		if len(b) > 0 {
			if err = ioutil.WriteFile(filename, b, 0644); err != nil {
				log.Println(err)
				continue
			}
			filenames = append(filenames, filename)
		} else {
			log.Println("No content, skipped", host.hostname)
		}
	}
	return filenames, err
}
