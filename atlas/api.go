// Copyright 2019 Kuei-chun Chen. All rights reserved.

package atlas

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/simagix/gox"
)

// BaseURL -
const BaseURL = "https://cloud.mongodb.com/api/atlas/v1.0"

// ApplicationJSON -
const ApplicationJSON = "application/json"

// ApplicationGZip -
const ApplicationGZip = "application/gzip"

// API stores Atlas API key
type API struct {
	acceptType  string
	alerts      bool
	alertsFile  string
	args        []string
	contentType string
	clusterName string
	ftdc        bool
	groupID     string
	info        bool
	loginfo     bool
	logNames    []string
	pause       bool
	resume      bool
	params      []string
	privateKey  string
	publicKey   string
	request     string
	verbose     bool
}

// NewKey returns API struct
func NewKey(publicKey string, privateKey string) *API {
	return &API{publicKey: publicKey, privateKey: privateKey, contentType: ApplicationJSON, acceptType: ApplicationJSON}
}

// ParseURI returns API struct from a URI
func ParseURI(uri string) (*API, error) {
	api := &API{contentType: ApplicationJSON, acceptType: ApplicationJSON, params: []string{}}
	uri = strings.TrimPrefix(uri, "atlas://")
	i := strings.Index(uri, "@")
	if i > 0 {
		tailer := uri[i+1:]
		if q := strings.Index(tailer, "?"); q > 0 {
			api.params = strings.Split(tailer[q+1:], "&")
			tailer = tailer[:q]
		}
		if n := strings.Index(tailer, "/"); n > 0 {
			api.groupID = tailer[:n]
			api.clusterName = tailer[n+1:]
		} else {
			api.groupID = tailer
		}
		uri = uri[:i]
	}
	i = strings.LastIndex(uri, ":")
	if i < 0 {
		return nil, errors.New("invalid format ([atlas://]publicKey:privateKey[@group/cluster])")
	}
	api.publicKey = uri[:i]
	api.privateKey = uri[i+1:]
	return api, nil
}

// GetLogNames gets downloaded log names
func (api *API) GetLogNames() []string {
	return api.logNames
}

// SetAcceptType sets acceptType
func (api *API) SetAcceptType(acceptType string) {
	api.acceptType = acceptType
}

// SetAlerts sets alerts
func (api *API) SetAlerts(alerts bool) {
	api.alerts = alerts
}

// SetAlertsFile sets alerts
func (api *API) SetAlertsFile(alertsFile string) {
	api.alertsFile = alertsFile
}

// SetArgs sets args
func (api *API) SetArgs(args []string) {
	api.args = args
}

// SetContentType sets contentType
func (api *API) SetContentType(contentType string) {
	api.contentType = contentType
}

// SetFTDC sets ftdc
func (api *API) SetFTDC(ftdc bool) {
	api.ftdc = ftdc
}

// SetInfo sets info
func (api *API) SetInfo(info bool) {
	api.info = info
}

// SetLoginfo sets loginfo
func (api *API) SetLoginfo(loginfo bool) {
	api.loginfo = loginfo
}

// SetPause sets pause
func (api *API) SetPause(pause bool) {
	api.pause = pause
}

// SetRequest sets request
func (api *API) SetRequest(request string) {
	api.request = request
}

// SetResume sets resume
func (api *API) SetResume(resume bool) {
	api.resume = resume
}

// SetVerbose sets verbose
func (api *API) SetVerbose(verbose bool) {
	api.verbose = verbose
}

// Get performs HTTP GET function
func (api *API) Get(uri string) ([]byte, error) {
	var err error
	var resp *http.Response
	var b []byte

	headers := map[string]string{}
	headers["Content-Type"] = api.contentType
	headers["Accept"] = api.acceptType
	if resp, err = gox.HTTPDigest("GET", uri, api.publicKey, api.privateKey, headers); err != nil {
		return b, err
	}
	defer resp.Body.Close()
	b, err = ioutil.ReadAll(resp.Body)
	return b, err
}

// Patch performs HTTP PATCH function
func (api *API) Patch(uri string, body []byte) ([]byte, error) {
	var err error
	var resp *http.Response
	var b []byte

	headers := map[string]string{}
	headers["Content-Type"] = api.contentType
	headers["Accept"] = api.acceptType
	if resp, err = gox.HTTPDigest("PATCH", uri, api.publicKey, api.privateKey, headers, body); err != nil {
		return b, err
	}
	defer resp.Body.Close()
	b, err = ioutil.ReadAll(resp.Body)
	return b, err
}

// Execute executes rest calls
func (api *API) Execute() string {
	var err error
	var str string
	if api.info {
		if str, err = api.GetClustersSummary(); err != nil {
			return err.Error()
		}
		return str
	} else if api.loginfo {
		if api.logNames, err = api.DownloadLogs(); err != nil {
			return err.Error()
		}
		if len(api.logNames) > 0 {
			return "Files downloaded:\n" + "\t" + strings.Join(api.logNames, "\n\t ")
		}
		return "No file downloaded"
	} else if api.ftdc {
		if str, err = api.DownloadFTDC(); err != nil {
			return err.Error()
		}
		return str
	} else if api.resume {
		if str, err = api.ClustersDo("PATCH", `{ "paused": false }`); err != nil {
			return err.Error()
		}
		return str
	} else if api.pause {
		if str, err = api.ClustersDo("PATCH", `{ "paused": true }`); err != nil {
			return err.Error()
		}
		return str
	} else if api.request != "" {
		data := "{}"
		if len(api.args) > 1 {
			data = api.args[1]
		}
		if api.request == "DELETE" {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Enter cluster name (", api.clusterName, "): ")
			text, _ := reader.ReadString('\n')
			text = strings.TrimRight(text, "\n")
			if text != api.clusterName {
				return "Cluster name does not match"
			}
		}
		if str, err = api.ClustersDo(api.request, data); err != nil {
			return err.Error()
		}
		return str
	} else if api.alerts {
		if str, err = api.AlertsDo("GET", `{ }`); err != nil {
			return err.Error()
		}
		return str
	} else if api.alertsFile != "" {
		if str, err = api.AddAlerts(api.alertsFile); err != nil {
			return err.Error()
		}
		return str
	}
	return "No argument included"
}
