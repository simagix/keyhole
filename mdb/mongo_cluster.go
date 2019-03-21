// Copyright 2019 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoCluster holds mongo cluster info
type MongoCluster struct {
	client       *mongo.Client
	hostname     string
	htmlFilename string
	jsonFilename string
	verbose      bool
}

// NewMongoCluster server info struct
func NewMongoCluster(client *mongo.Client) *MongoCluster {
	return &MongoCluster{client: client}
}

// SetHostname -
func (mc *MongoCluster) SetHostname(hostname string) {
	mc.hostname = hostname
	mc.htmlFilename = hostname + ".html"
	mc.jsonFilename = hostname + ".json"
}

// SetVerbose -
func (mc *MongoCluster) SetVerbose(verbose bool) {
	mc.verbose = verbose
}

// GetInfo -
func (mc *MongoCluster) GetInfo() (bson.M, error) {
	var err error
	var cur *mongo.Cursor
	var icur *mongo.Cursor
	var ctx = context.Background()
	var config = bson.M{}
	var cluster = bson.M{"config": config}
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
	cluster["cluster"] = info.Cluster
	cluster["host"] = info.Host
	cluster["process"] = info.Process
	cluster["sharding"] = info.Sharding
	cluster["storage"] = info.StorageSize
	cluster["version"] = info.Version
	// hostInfo
	var hostInfo bson.M
	if hostInfo, err = RunAdminCommand(mc.client, "hostInfo"); err != nil {
		return cluster, err
	}
	config["hostInfo"] = trimMap(hostInfo)

	// getCmdLineOpts
	var getCmdLineOpts bson.M
	if getCmdLineOpts, err = RunAdminCommand(mc.client, "getCmdLineOpts"); err != nil {
		return cluster, err
	}
	config["getCmdLineOpts"] = trimMap(getCmdLineOpts)

	// buildInfo
	var buildInfo bson.M
	if buildInfo, err = RunAdminCommand(mc.client, "buildInfo"); err != nil {
		return cluster, err
	}
	config["buildInfo"] = trimMap(buildInfo)

	// ServerStatus
	var serverStatus bson.M
	if serverStatus, err = RunAdminCommand(mc.client, "serverStatus"); err != nil {
		return cluster, err
	}
	config["serverStatus"] = trimMap(serverStatus)

	// replSetGetStatus
	if info.Cluster == "replica" {
		var replSetGetStatus bson.M
		if replSetGetStatus, err = RunAdminCommand(mc.client, "replSetGetStatus"); err != nil {
			return cluster, err
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
			return cluster, err
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
				val := bson.M{}
				icur.Decode(&val)
				indexes = append(indexes, val)
			}
			icur.Close(ctx)

			// stats
			var stats bson.M
			mc.client.Database(dbName).RunCommand(ctx, bson.D{{Key: "collStats", Value: collectionName}}).Decode(&stats)
			delete(stats, "indexDetails")
			delete(stats, "wiredTiger")
			collections = append(collections, bson.M{"NS": ns, "collection": collectionName, "document": firstDoc, "indexes": indexes, "stats": trimMap(stats)})
		}
		var stats bson.M
		stats, _ = RunCommandOnDB(mc.client, "dbStats", dbName)
		databases = append(databases, bson.M{"DB": dbName, "collections": collections, "stats": trimMap(stats)})
	}
	cluster["databases"] = databases
	if mc.hostname != "" {
		b := []byte(Stringify(cluster, "", "  "))
		ioutil.WriteFile(mc.jsonFilename, b, 0644)
		if err = mc.outputHTML(cluster); err != nil {
			fmt.Println("Error writing", mc.htmlFilename, err)
		}
	}
	return cluster, err
}

func (mc *MongoCluster) outputHTML(cluster bson.M) error {
	var err error
	var b []byte
	var strs []string

	strs = append(strs, "<h1>3 Cluster Data Stats</h1>")
	counter := 0
	for _, database := range cluster["databases"].([]bson.M) {
		db := database["DB"].(string)
		if db == "admin" || db == "local" || db == "config" {
			continue
		}
		for _, collection := range database["collections"].([]bson.M) {
			counter++
			cnt := fmt.Sprintf("%d", counter)
			stats := collection["stats"].(bson.M)
			indexes := collection["indexes"].([]bson.M)
			// collection stats
			strs = append(strs, "<h2>3."+cnt+" Collection "+collection["NS"].(string)+"</h2>")
			strs = append(strs, "<h3>3."+cnt+".1 Stats</h3>")
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
			strs = append(strs, "<td align=right>"+fmt.Sprintf("%v", getSize(stats["count"]))+"</td>")
			strs = append(strs, "</tr>")
			strs = append(strs, "<tr>")
			strs = append(strs, `<td class="rowtitle">Average Document Size</td>`)
			strs = append(strs, "<td align=right>"+fmt.Sprintf("%v", getSize(stats["avgObjSize"]))+"</td>")
			strs = append(strs, "</tr>")
			strs = append(strs, "<tr>")
			strs = append(strs, `<td class="rowtitle">Storage Size</td>`)
			strs = append(strs, "<td align=right>"+fmt.Sprintf("%v", getSize(stats["storageSize"]))+"</td>")
			strs = append(strs, "</tr>")
			strs = append(strs, "<tr>")
			strs = append(strs, `<td class="rowtitle">Data Size</td>`)
			strs = append(strs, "<td align=right>"+fmt.Sprintf("%v", getSize(stats["size"]))+"</td>")
			strs = append(strs, "</tr>")
			strs = append(strs, "<tr>")
			strs = append(strs, `<td class="rowtitle">Total Indexes Size</td>`)
			strs = append(strs, "<td align=right>"+fmt.Sprintf("%v", getSize(stats["totalIndexSize"]))+"</td>")
			strs = append(strs, "</tr>")
			strs = append(strs, " </tbody>")
			strs = append(strs, "</table><br/>")

			// indexes
			strs = append(strs, "<h3>3."+cnt+".2 Indexes</h3>")
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
				strs = append(strs, "<td align=right>"+fmt.Sprintf("%v", getSize(stats["indexSizes"].(bson.M)[name]))+"</td>")
				strs = append(strs, "</tr>")
			}

			strs = append(strs, " </tbody>")
			strs = append(strs, "</table><br/>")

			strs = append(strs, "<h3>3."+cnt+".3 Sample Document</h3>")
			strs = append(strs, "<pre>")
			strs = append(strs, Stringify(collection["document"], "", "  "))
			strs = append(strs, "</pre><br/>")
		}
	}
	htmlTemplate = strings.Replace(htmlTemplate, "{TODO}", strings.Join(strs, "\n"), -1)
	b = []byte(htmlTemplate)
	err = ioutil.WriteFile(mc.htmlFilename, b, 0644)
	return err
}

func trimMap(doc bson.M) bson.M {
	delete(doc, "$clusterTime")
	delete(doc, "operationTime")
	delete(doc, "ok")
	return doc
}

func getSize(num interface{}) string {
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
