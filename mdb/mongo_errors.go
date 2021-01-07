// Copyright 2020 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
)

// IsUnauthorizedError check Unauthorized error
func IsUnauthorizedError(err error) bool {
	e, ok := err.(mongo.CommandError)
	return ok && e.Code == 13
}

// GetErrorCode gets error code for debug purpose
func GetErrorCode(err error) int {
	switch e := err.(type) {
	case mongo.CommandError:
		return int(e.Code)
	case mongo.BulkWriteError:
		return e.Code
	case mongo.BulkWriteException:
		fmt.Println("BulkWriteException")
		return 0
	case mongo.WriteError:
		return e.Code
	case mongo.WriteException:
		fmt.Println("WriteException")
		return 0
	default:
		fmt.Println("unknown type") // prints unknown error type
		return 0
	}

}
