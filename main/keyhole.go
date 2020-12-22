// Copyright 2020 Kuei-chun Chen. All rights reserved.

package main

import (
	"fmt"
	"time"

	"github.com/simagix/keyhole"
)

var repo = "simagix/keyhole"
var version = "devel-xxxxxx"

func main() {
	if version == "devel-xxxxxx" {
		version = "devel-" + time.Now().Format("20060102")
	}
	fullVersion := fmt.Sprintf(`%v %v`, repo, version)
	keyhole.Run(fullVersion)
}
