// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"encoding/json"
)

// Silent does nothing
func Silent(data interface{}) {
}

// Stringify return a formatted string from JSON
func Stringify(doc interface{}, opts ...string) string {
	if len(opts) == 2 {
		b, _ := json.MarshalIndent(doc, opts[0], opts[1])
		return string(b)
	}
	b, _ := json.Marshal(doc)
	return string(b)
}
