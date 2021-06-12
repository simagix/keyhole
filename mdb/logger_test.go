// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"testing"
)

func TestInfo(t *testing.T) {
	logger := NewLogger("TestInfo", "-test")
	logger.Info("first ", "second ", 3, 4)
}

func TestInfof(t *testing.T) {
	logger := NewLogger("TestInfof", "-test")
	logger.Infof(`%v => %v`, "key", 100)
}

func TestErrorf(t *testing.T) {
	logger := NewLogger("TestErrorf", "-test")
	logger.Errorf(`%v => %v`, "key", 100)
}

func TestWarnf(t *testing.T) {
	logger := NewLogger("TestWarnf", "-test")
	logger.Warnf(`%v => %v`, "key", 100)
}

func TestDebug(t *testing.T) {
	logger := GetLogger("TestDebug")
	logger.Debug(`no message`)
	logger.SetLoggerLevel(Debug)
	logger.Debug(`debug message`)
	logger.SetLoggerLevel(Info)
	logger.Debug(`no message`)
	logger.SetLoggerLevel(Trace)
	logger.Debug(`trace message`)
}
