// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mongo

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// RunAdminCommand executes admin Command
func RunAdminCommand(session *mgo.Session, command string) (bson.M, error) {
	return RunCommandOnDB(session, command, "admin")
}

// RunCommandOnDB execute admin Command at given database
func RunCommandOnDB(session *mgo.Session, command string, db string) (bson.M, error) {
	result := bson.M{}
	if err := session.DB(db).Run(command, &result); err != nil {
		return nil, err
	}
	return result, nil
}
