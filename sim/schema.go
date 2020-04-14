// Copyright 2018 Kuei-chun Chen. All rights reserved.

package sim

import (
	"context"
	"encoding/json"
	"regexp"

	"github.com/simagix/gox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetSchema returns a masked first doc of a collection
func GetSchema(c *mongo.Collection, verbose bool) (string, error) {
	var err error
	var buf []byte
	var doc bson.M
	if err = c.FindOne(context.Background(), bson.M{}).Decode(&doc); err != nil {
		return "", err
	}
	if buf, err = bson.MarshalExtJSON(doc, false, false); err != nil {
		return "", err
	}
	json.Unmarshal(buf, &doc)
	str := gox.Stringify(doc, "", "  ")
	if verbose == true {
		return str, err
	}
	re := regexp.MustCompile(`{\s+"\$oid":\s?("[a-fA-F0-9]{24}")\s+}`)
	str = re.ReplaceAllString(str, "ObjectId($1)")
	re = regexp.MustCompile(`{\s+"\$date":\s?("\S+")\s+}`)
	str = re.ReplaceAllString(str, "ISODate($1)")
	re = regexp.MustCompile(`{\s+"\$numberDecimal":\s?("\S+")\s+}`)
	str = re.ReplaceAllString(str, "NumberDecimal($1)")
	re = regexp.MustCompile(`{\s+"\$numberDouble":\s?("NaN")\s+}`)
	str = re.ReplaceAllString(str, "NaN")
	return str, err
}
