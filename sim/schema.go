// Copyright 2018 Kuei-chun Chen. All rights reserved.

package sim

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	"github.com/simagix/gox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetSchema returns a masked first doc of a collection
func GetSchema(c *mongo.Collection, verbose bool) (string, error) {
	var err error
	var data []byte
	var doc bson.M
	if err = c.FindOne(context.Background(), bson.M{}).Decode(&doc); err != nil {
		return "", err
	}
	if data, err = bson.MarshalExtJSON(doc, false, false); err != nil {
		return "", err
	}
	if verbose {
		os.Mkdir(outdir, 0755)
		ofile := fmt.Sprintf("%v/%v.json", outdir, c.Name())
		err = os.WriteFile(ofile, data, 0644)
		fmt.Println(err, ofile, string(data))
		return fmt.Sprintf("json template written to %v", ofile), err
	}
	str := string(data)
	re := regexp.MustCompile(`{"\$binary":{"base64":"[^"]*","subType":"(0[0-3])"}}`)
	str = re.ReplaceAllString(str, `{"$$binary":{"base64":"","subType":"$1"}}`)
	re = regexp.MustCompile(`{"\$binary":{"base64":"[^"]{24}","subType":"04"}}`)
	str = re.ReplaceAllString(str, `"$$uuid"`)
	re = regexp.MustCompile(`{"\$binary":{"base64":"","subType":"0([0-3])"}}`)
	str = re.ReplaceAllString(str, `"BinData($1, '')"`)
	json.Unmarshal([]byte(str), &doc)
	str = gox.Stringify(doc, "", "  ")
	re = regexp.MustCompile(`{\s+"\$oid":\s?("[a-fA-F0-9]{24}")\s+}`)
	str = re.ReplaceAllString(str, "ObjectId($1)")
	re = regexp.MustCompile(`{\s+"\$date":\s?("\S+")\s+}`)
	str = re.ReplaceAllString(str, "ISODate($1)")
	re = regexp.MustCompile(`{\s+"\$numberDecimal":\s?("\S+")\s+}`)
	str = re.ReplaceAllString(str, "NumberDecimal($1)")
	re = regexp.MustCompile(`{\s+"\$numberDouble":\s?("NaN")\s+}`)
	str = re.ReplaceAllString(str, "NaN")
	re = regexp.MustCompile(`"\$uuid"`)
	str = re.ReplaceAllString(str, gox.GetRandomUUIDString())
	re = regexp.MustCompile(`"(BinData\(\d, ''\))"`)
	str = re.ReplaceAllString(str, `$1`)
	return str, err
}
