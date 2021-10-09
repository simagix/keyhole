// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"reflect"
	"regexp"
	"strings"
)

// Redactor stores Redact struct
type Redactor struct {
	verbose bool
}

// NewRedactor returns  Redact struct
func NewRedactor() *Redactor { return &Redactor{} }

// SetVerbose sets verbosity
func (r *Redactor) SetVerbose(verbose bool) {
	r.verbose = verbose
}

func (r *Redactor) callback(v interface{}) interface{} {
	if v == nil {
		return v
	}
	vt := reflect.TypeOf(v)
	switch vt.Kind() {
	case reflect.Bool:
		return false
	case reflect.String:
		str := v.(string)
		var mdate = regexp.MustCompile(`^\d{4}\-(0?[1-9]|1[012])\-(0?[1-9]|[12][0-9]|3[01])T.*$`)
		if mdate.MatchString(str) {
			return "1970-01-01T" + str[11:]
		}
		hdr := ""
		if n := strings.Index(str, "://"); n > 0 {
			hdr = str[:n+3]
		}
		b := make([]byte, len(str))
		for i, c := range str {
			if i < len(hdr) {
				b[i] = byte(c)
			} else if c >= 48 && c <= 57 { // digits
				b[i] = byte(1 + 48) // 1
			} else if c >= 65 && c <= 70 { // A-F
				b[i] = byte(5 + 65) // F
			} else if c >= 71 && c <= 90 { // G-Z
				b[i] = byte(23 + 65) // X
			} else if c >= 97 && c <= 122 { // a-f
				b[i] = byte(5 + 97) // f
			} else if c >= 103 && c <= 142 { // g-z
				b[i] = byte(23 + 97) // x
			} else {
				b[i] = byte(c)
			}
		}
		return string(b)
	default:
		return v
	}
}
