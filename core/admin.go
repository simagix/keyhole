// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// AdminCommand executes admin Command
func AdminCommand(session *mgo.Session, command string) (bson.M, error) {
	return AdminCommandOnDB(session, command, "admin")
}

// AdminCommandOnDB execute admin Command at given database
func AdminCommandOnDB(session *mgo.Session, command string, db string) (bson.M, error) {
	session.SetMode(mgo.Primary, true)
	result := bson.M{}
	if err := session.DB(db).Run(command, &result); err != nil {
		return nil, err
	}
	return result, nil
}
