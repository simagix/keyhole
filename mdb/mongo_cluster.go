// Copyright 2019 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// MongoCluster holds mongo cluster info
type MongoCluster struct {
	client  *mongo.Client
	cluster bson.M
	verbose bool
}

// NewMongoCluster server info struct
func NewMongoCluster(client *mongo.Client) *MongoCluster {
	return &MongoCluster{client: client}
}

// SetVerbose -
func (mc *MongoCluster) SetVerbose(verbose bool) {
	mc.verbose = verbose
}

// GetClusterInfo -
func (mc *MongoCluster) GetClusterInfo() (bson.M, error) {
	var err error
	var cur *mongo.Cursor
	var icur *mongo.Cursor
	var scur *mongo.Cursor
	var ctx = context.Background()
	var config = bson.M{}
	var pipeline = MongoPipeline(`{"$indexStats": {}}`)

	mc.cluster = bson.M{"config": config}
	var info ServerInfo
	if info, err = GetServerInfo(mc.client); err != nil {
		return nil, err
	}
	var val bson.M
	b, _ := bson.Marshal(info)
	bson.Unmarshal(b, &val)
	if mc.verbose == false {
		return val, err
	}

	info.StorageSize["databases"] = info.StorageSize["statsDetails"]
	delete(info.StorageSize, "statsDetails")
	mc.cluster["cluster"] = info.Cluster
	mc.cluster["host"] = info.Host
	mc.cluster["process"] = info.Process
	mc.cluster["sharding"] = info.Sharding
	mc.cluster["storage"] = info.StorageSize
	mc.cluster["version"] = info.Version
	// hostInfo
	var hostInfo bson.M
	if hostInfo, err = RunAdminCommand(mc.client, "hostInfo"); err != nil {
		return mc.cluster, err
	}
	config["hostInfo"] = trimMap(hostInfo)

	// getCmdLineOpts
	var getCmdLineOpts bson.M
	if getCmdLineOpts, err = RunAdminCommand(mc.client, "getCmdLineOpts"); err != nil {
		return mc.cluster, err
	}
	config["getCmdLineOpts"] = trimMap(getCmdLineOpts)

	// buildInfo
	var buildInfo bson.M
	if buildInfo, err = RunAdminCommand(mc.client, "buildInfo"); err != nil {
		return mc.cluster, err
	}
	config["buildInfo"] = trimMap(buildInfo)

	// ServerStatus
	var serverStatus bson.M
	if serverStatus, err = RunAdminCommand(mc.client, "serverStatus"); err != nil {
		return mc.cluster, err
	}
	config["serverStatus"] = trimMap(serverStatus)

	// replSetGetStatus
	if info.Cluster == "replica" {
		var replSetGetStatus bson.M
		if replSetGetStatus, err = RunAdminCommand(mc.client, "replSetGetStatus"); err != nil {
			return mc.cluster, err
		}
		config["replSetGetStatus"] = trimMap(replSetGetStatus)
	}

	// usersInfo
	var usersInfo bson.M
	if usersInfo, err = RunAdminCommand(mc.client, "usersInfo"); err != nil {
		log.Println(err)
	}
	config["usersInfo"] = trimMap(usersInfo)

	// rolesInfo
	var rolesInfo bson.M
	if rolesInfo, err = RunAdminCommand(mc.client, "rolesInfo"); err != nil {
		log.Println(err)
	}
	config["rolesInfo"] = trimMap(rolesInfo)

	// collections firstDoc (findOne), indexes, and stats
	dbNames, _ := ListDatabaseNames(mc.client)
	var databases = []bson.M{}
	for _, dbName := range dbNames {
		// if dbName == "admin" || dbName == "config" || dbName == "local" {
		// 	continue
		// }
		if cur, err = mc.client.Database(dbName).ListCollections(ctx, bson.M{}); err != nil {
			return mc.cluster, err
		}
		defer cur.Close(ctx)
		var collections = []bson.M{}

		for cur.Next(ctx) {
			var elem = bson.M{}
			if err = cur.Decode(&elem); err != nil {
				continue
			}
			collectionName := fmt.Sprintf("%v", elem["name"])
			ns := dbName + "." + collectionName
			collection := mc.client.Database(dbName).Collection(collectionName)

			// firstDoc, FindOne
			var firstDoc bson.M
			if err = collection.FindOne(ctx, bson.D{{}}).Decode(&firstDoc); err != nil {
				continue
			}

			// indexes
			view := collection.Indexes()
			if icur, err = view.List(ctx); err != nil {
				continue
			}

			indexes := []bson.M{}
			for icur.Next(ctx) {
				idx := bson.D{}
				icur.Decode(&idx)
				val := bson.M{}
				icur.Decode(&val)
				val["stats"] = []bson.M{}

				var strbuf bytes.Buffer
				var keys bson.D

				for _, v := range idx {
					if v.Key == "key" {
						keys = v.Value.(bson.D)
					}
				}

				for n, value := range keys {
					if n == 0 {
						strbuf.WriteString("{ ")
					}
					strbuf.WriteString(value.Key + ": " + fmt.Sprint(value.Value))
					if n == len(keys)-1 {
						strbuf.WriteString(" }")
					} else {
						strbuf.WriteString(", ")
					}
				}
				keystr := strbuf.String()
				val["effectiveKey"] = strings.Replace(keystr[2:len(keystr)-2], ": -1", ": 1", -1)
				indexes = append(indexes, val)
			}
			icur.Close(ctx)

			// stats
			var stats bson.M
			mc.client.Database(dbName).RunCommand(ctx, bson.D{{Key: "collStats", Value: collectionName}}).Decode(&stats)
			delete(stats, "indexDetails")
			delete(stats, "wiredTiger")

			if dbName != "admin" && dbName != "local" && dbName != "config" {
				if scur, err = mc.client.Database(dbName).Collection(collectionName).Aggregate(ctx, pipeline); err != nil {
					log.Fatal(dbName, err)
				}
				for scur.Next(ctx) {
					var result = bson.M{}
					if err = scur.Decode(&result); err != nil {
						continue
					}
					for _, index := range indexes {
						if index["name"] == result["name"] {
							delete(result, "key")
							delete(result, "name")
							index["stats"] = append(index["stats"].([]bson.M), result)
							break
						}
					}
				}
				scur.Close(ctx)
			}
			collections = append(collections, bson.M{"NS": ns, "collection": collectionName, "document": firstDoc,
				"indexes": indexes, "stats": trimMap(stats)})
		}
		var stats bson.M
		stats, _ = RunCommandOnDB(mc.client, "dbStats", dbName)
		databases = append(databases, bson.M{"DB": dbName, "collections": collections, "stats": trimMap(stats)})
	}
	mc.cluster["databases"] = databases
	return mc.cluster, err
}

// WriteJSON outputs cluster into to a JSON file
func (mc *MongoCluster) WriteJSON(filename string) error {
	b := []byte(Stringify(mc.cluster))
	return ioutil.WriteFile(filename, b, 0644)
}

// WriteHTML outputs cluster into to a HTML file
func (mc *MongoCluster) WriteHTML(filename string) error {
	html, _ := mc.getClusterHTML()
	b := []byte(html)
	return ioutil.WriteFile(filename, b, 0644)
}

// getClusterHTML -
func (mc *MongoCluster) getClusterHTML() (string, error) {
	var err error
	var strs []string
	var toc []string

	toc = append(toc, "<a name=toc></a>")
	toc = append(toc, "<h2>TOC</h2><ul>")
	strs = append(strs, "<h1>3 Cluster Data Stats</h1>")
	p := message.NewPrinter(language.English)
	counter := 0
	for _, database := range mc.cluster["databases"].([]bson.M) {
		db := database["DB"].(string)
		if db == "admin" || db == "local" || db == "config" {
			continue
		}
		for _, collection := range database["collections"].([]bson.M) {
			counter++
			tag := fmt.Sprintf("3.%d", counter)
			stats := collection["stats"].(bson.M)
			indexes := collection["indexes"].([]bson.M)
			// collection stats
			toc = append(toc, "<li><a href=#"+tag+">"+tag+" Collection "+collection["NS"].(string)+"</a></li>")
			strs = append(strs, "<a name="+tag+"></a>")
			strs = append(strs, "<a href=#toc>top</a>")
			strs = append(strs, "<h2>"+tag+" Collection "+collection["NS"].(string)+"</h2>")
			strs = append(strs, "<h3>"+tag+".1 Stats</h3>")
			strs = append(strs, "<table>")
			strs = append(strs, "<thead>")
			strs = append(strs, " <tr>")
			strs = append(strs, `<th width="50%">Stats</th>`)
			strs = append(strs, `<th width="50%">Total</th>`)
			strs = append(strs, "</tr>")
			strs = append(strs, "</thead>")
			strs = append(strs, " <tbody>")
			strs = append(strs, "<tr>")
			strs = append(strs, `<td class="rowtitle">Number of Documents</td>`)
			strs = append(strs, "<td align=right>"+fmt.Sprintf("%v", p.Sprintf("%d", stats["count"].(int32)))+"</td>")
			strs = append(strs, "</tr>")
			strs = append(strs, "<tr>")
			strs = append(strs, `<td class="rowtitle">Average Document Size</td>`)
			strs = append(strs, "<td align=right>"+fmt.Sprintf("%v", GetStorageSize(stats["avgObjSize"]))+"</td>")
			strs = append(strs, "</tr>")
			strs = append(strs, "<tr>")
			strs = append(strs, `<td class="rowtitle">Storage Size</td>`)
			strs = append(strs, "<td align=right>"+fmt.Sprintf("%v", GetStorageSize(stats["storageSize"]))+"</td>")
			strs = append(strs, "</tr>")
			strs = append(strs, "<tr>")
			strs = append(strs, `<td class="rowtitle">Data Size</td>`)
			strs = append(strs, "<td align=right>"+fmt.Sprintf("%v", GetStorageSize(stats["size"]))+"</td>")
			strs = append(strs, "</tr>")
			strs = append(strs, "<tr>")
			strs = append(strs, `<td class="rowtitle">Total Indexes Size</td>`)
			strs = append(strs, "<td align=right>"+fmt.Sprintf("%v", GetStorageSize(stats["totalIndexSize"]))+"</td>")
			strs = append(strs, "</tr>")
			strs = append(strs, " </tbody>")
			strs = append(strs, "</table><br/>")

			// indexes
			strs = append(strs, "<h3>"+tag+".2 Indexes</h3>")
			strs = append(strs, "<table>")
			strs = append(strs, "<thead>")
			strs = append(strs, " <tr>")
			strs = append(strs, `<th width="42%">Name</th>`)
			strs = append(strs, `<th width="42%">Definition</th>`)
			strs = append(strs, `<th width="16%">Size</th>`)
			strs = append(strs, "</tr>")
			strs = append(strs, "</thead>")
			strs = append(strs, " <tbody>")

			for _, index := range indexes {
				name := index["name"].(string)
				strs = append(strs, "<tr>")
				strs = append(strs, `<td>`+name+`</td>`)
				strs = append(strs, `<td>`+Stringify(index["key"])+`</td>`)
				strs = append(strs, "<td align=right>"+fmt.Sprintf("%v", GetStorageSize(stats["indexSizes"].(bson.M)[name]))+"</td>")
				strs = append(strs, "</tr>")
			}

			strs = append(strs, " </tbody>")
			strs = append(strs, "</table><br/>")

			strs = append(strs, "<h3>"+tag+".3 Sample Document</h3>")
			strs = append(strs, "<pre>")
			strs = append(strs, Stringify(collection["document"], "", "  "))
			strs = append(strs, "</pre><br/>")
		}
	}
	toc = append(toc, "</ul>")
	toc = append(toc, strs...)
	htmlTemplate = strings.Replace(htmlTemplate, "{TODO}", strings.Join(toc, ""), -1)
	return htmlTemplate, err
}

func trimMap(doc bson.M) bson.M {
	delete(doc, "$clusterTime")
	delete(doc, "operationTime")
	delete(doc, "ok")
	return doc
}

// GetStorageSize returns storage size in [TGMK] B
func GetStorageSize(num interface{}) string {
	f := fmt.Sprintf("%v", num)
	x, err := strconv.ParseFloat(f, 64)
	if err != nil {
		return f
	}

	if x >= (1024 * 1024 * 1024 * 1024) {
		s := fmt.Sprintf("%v", x/(1024*1024*1024*1024))
		return round(s) + " TB"
	} else if x >= (1024 * 1024 * 1024) {
		s := fmt.Sprintf("%v", x/(1024*1024*1024))
		return round(s) + " GB"
	} else if x >= (1024 * 1024) {
		s := fmt.Sprintf("%v", x/(1024*1024))
		return round(s) + " MB"
	} else if x >= 1024 {
		s := fmt.Sprintf("%v", x/1024)
		return round(s) + " KB"
	}
	return fmt.Sprintf("%v B", x)
}

func round(s string) string {
	idx := strings.Index(s, ".")
	if idx < 0 {
		return s
	}
	return s[:idx+2]
}

var htmlTemplate = `
<!DOCTYPE html>
<html lang="en">

<head>
  <title>MongoDB Cluster Info</title>
	<meta http-equiv="Cache-Control" content="no-cache, no-store, must-revalidate" />
	<meta http-equiv="Pragma" content="no-cache" />
	<meta http-equiv="Expires" content="0" />
  <style>
		body {
			font-family: Arial, Helvetica, sans-serif;
		}
    table
    {
			font-family: Arial, Helvetica, sans-serif;
    	border-collapse:collapse;
    	width:100%;
    	max-width:700px;
    	min-width:400px;
    	#text-align:center;
    }
    caption
    {
    	caption-side:top;
    	font-weight:bold;
    	font-style:italic;
    	margin:4px;
    }
    table,th, td
    {
    	border: 1px solid gray;
    }
    th, td
    {
    	height: 24px;
    	padding:4px;
    	vertical-align:middle;
    }
    th
    {
    	#background-image:url(table-shaded.png);
      background-color: #000;
      color: #FFF;
    }
    .rowtitle
    {
    	font-weight:bold;
    }
		ul {
		  list-style-type: none;
		  margin: 0;
		  padding: 0;
		}
		li {
			padding-left: 20px;
		}
		li:last-child {
		  border: none;
		}
		a {
		  text-decoration: none;
		  color: #000;
		  display: block;

		  -webkit-transition: font-size 0.3s ease, background-color 0.3s ease;
		  -moz-transition: font-size 0.3s ease, background-color 0.3s ease;
		  -o-transition: font-size 0.3s ease, background-color 0.3s ease;
		  -ms-transition: font-size 0.3s ease, background-color 0.3s ease;
		  transition: font-size 0.3s ease, background-color 0.3s ease;
		}
		a:hover {
			color: blue;
		}
    </style>
</head>

<body>
  <script type="text/javascript">
    function colourize() {
      var dnl = document.getElementsByTagName("tr");

      for (i = 0; i < dnl.length; i++) {
        if ((Math.round(i / 2) * 2) == ((i / 2) * 2))
          dnl.item(i).style.background = "#F0F0F0";
      }
    }

    window.onload = colourize;
  </script>
	{TODO}
</body>
`
