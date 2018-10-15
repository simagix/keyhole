// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"log"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// AdminCommand executes admin Command
func AdminCommand(session *mgo.Session, command string) (bson.M, error) {
	return AdminCommandOnDB(session, command, "admin")
}

// AdminCommandOnDB execute admin Command at given database
func AdminCommandOnDB(session *mgo.Session, command string, db string) (bson.M, error) {
	result := bson.M{}
	if err := session.DB(db).Run(command, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Cleanup drops the temp database
func Cleanup(session *mgo.Session) error {
	var err error
	log.Println("dropping collection", SimDBName, CollectionName)
	session.DB(SimDBName).C(CollectionName).DropCollection()
	log.Println("dropping database", SimDBName)
	session.DB(SimDBName).DropDatabase()
	return err
}
