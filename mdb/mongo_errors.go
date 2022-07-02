// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"fmt"

	"go.mongodb.org/mongo-driver/x/mongo/driver/topology"

	"go.mongodb.org/mongo-driver/mongo"
)

const (
	// TopologyConnectionError keyhole custom error codes
	TopologyConnectionError int   = 20001
	unauthorizedError       int32 = 13
)

var (
	duplicatedKeyErrorCodes = []int{11000, 11001}
)

// IsDuplicateKeyError check if error is from duplicate key
func IsDuplicateKeyError(err error) bool {
	werr, ok := err.(mongo.WriteError)
	if ok {
		return isDuplicateKeyCode(werr.Code)
	}
	blkerr, ok := err.(mongo.BulkWriteError)
	if ok {
		return isDuplicateKeyCode(blkerr.Code)
	}
	we, ok := err.(mongo.WriteException)
	if ok {
		return isDuplicateKeyCode(GetErrorCode(we))
	}
	blke, ok := err.(mongo.BulkWriteException)
	return ok && isDuplicateKeyCode(GetErrorCode(blke))
}

// IsUnauthorizedError check Unauthorized error
func IsUnauthorizedError(err error) bool {
	e, ok := err.(mongo.CommandError)
	return ok && e.Code == unauthorizedError
}

// GetErrorCode gets error code for debug purpose
func GetErrorCode(err error) int {
	switch e := err.(type) {
	case mongo.CommandError:
		return int(e.Code)
	case mongo.BulkWriteError:
		fmt.Println("BulkWriteError")
		return e.Code
	case mongo.BulkWriteException:
		if len(e.WriteErrors) > 0 {
			return e.WriteErrors[0].Code
		}
		if e.WriteConcernError != nil {
			return e.WriteConcernError.Code
		}
		return 0
	case mongo.WriteError:
		return e.Code
	case mongo.WriteException:
		if len(e.WriteErrors) > 0 {
			return e.WriteErrors[0].Code
		}
		if e.WriteConcernError != nil {
			return e.WriteConcernError.Code
		}
		return 0
	case topology.ConnectionError:
		return TopologyConnectionError
	default:
		fmt.Printf("unsupported type %T, %v\n", err, err) // prints unsupported error type
		return 0
	}

}

func isDuplicateKeyCode(code int) bool {
	for _, c := range duplicatedKeyErrorCodes {
		if code == c {
			return true
		}
	}
	return false
}
