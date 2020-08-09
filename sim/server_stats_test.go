// Copyright 2018 Kuei-chun Chen. All rights reserved.

package sim

import (
	"fmt"
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

func TestCollectServerStatus(t *testing.T) {
	// var channel = make(chan string)
	if os.Getenv("DATABASE_URL") != "" {
		UnitTestURL = os.Getenv("DATABASE_URL")
	}
	connString, _ := connstring.Parse(UnitTestURL)
	runner, _ := NewRunner(connString)
	fmt.Println(runner.uri)
	// runner.CollectServerStatus(runner.uri, channel)
}
