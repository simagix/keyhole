// Copyright 2018 Kuei-chun Chen. All rights reserved.

package util

import (
	"errors"
	"fmt"
	"runtime"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

// ReadPasswordFromStdin reads password from stdin
func ReadPasswordFromStdin() (string, error) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		return "", errors.New("Missing password")
	}
	var buffer []byte
	var err error
	fmt.Print("Enter Password: ")
	if buffer, err = terminal.ReadPassword(int(syscall.Stdin)); err != nil {
		return "", err
	}
	return string(buffer), err
}
