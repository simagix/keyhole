// Copyright 2018 Kuei-chun Chen. All rights reserved.

package sim

import (
	"fmt"
	"os"
	"testing"
)

func TestCollectServerStatus(t *testing.T) {
	// var channel = make(chan string)
	if os.Getenv("DATABASE_URL") != "" {
		UnitTestURL = os.Getenv("DATABASE_URL")
	}
	runner, _ := NewRunner(UnitTestURL, "", "")
	fmt.Println(runner.uri)
	// runner.CollectServerStatus(runner.uri, channel)
}
