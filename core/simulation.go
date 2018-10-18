// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"encoding/json"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

var simDocs []bson.M

// initialize an array of documents for simulation test.  If a template is available
// read the sample json and replace them with random values.  Otherwise, use the demo
// example.
func (b Base) initSimDocs() {
	var err error
	var sdoc bson.M

	if b.verbose {
		log.Println("initSimDocs")
	}
	rand.Seed(time.Now().Unix())
	total := 512
	if b.filename == "" {
		for len(simDocs) < total {
			simDocs = append(simDocs, GetDemoDoc())
		}
		return
	}

	if sdoc, err = GetDocByTemplate(b.filename, true); err != nil {
		return
	}
	bytes, _ := json.MarshalIndent(sdoc, "", "   ")
	if b.verbose {
		log.Println(string(bytes))
	}
	doc := make(map[string]interface{})
	json.Unmarshal(bytes, &doc)

	for len(simDocs) < total {
		ndoc := make(map[string]interface{})
		RandomizeDocument(&ndoc, doc, false)
		delete(ndoc, "_id")
		ndoc["_search"] = strconv.FormatInt(rand.Int63(), 16)
		simDocs = append(simDocs, ndoc)
	}
}

// PopulateData - Insert docs to evaluate performance/bandwidth
// {
//	favorites: {
//		sports: []
//		cities: []
//	}
//	favoriteSports: []
//	favoriteSports1
//	favoriteSports2
//	favoriteSports3
// }
func (b Base) PopulateData(wmajor bool) error {
	var session *mgo.Session
	var err error

	if b.verbose {
		log.Println("PopulateData", wmajor)
	}
	if session, err = GetSession(b.dialInfo, wmajor, b.ssl, b.sslCAFile, b.sslPEMKeyFile); err != nil {
		return err
	}
	defer session.Close()
	c := session.DB(SimDBName).C(CollectionName)
	btime := time.Now()
	for time.Now().Sub(btime) < time.Minute {
		bulk := c.Bulk()
		var contentArray []interface{}
		docidx := 0
		for i := 0; i < b.bulkSize; i++ {
			contentArray = append(contentArray, simDocs[docidx%len(simDocs)])
			docidx++
		}
		bulk.Insert(contentArray...)
		if _, err = bulk.Run(); err != nil {
			return err
		}
	}

	return nil
}

// Simulate simulates CRUD for load tests
func (b Base) Simulate(duration int, transactions []Transaction, wmajor bool) {
	if b.verbose {
		log.Println("Simulate", duration, transactions, wmajor)
	}
	isTeardown := false
	var totalTPS int

	var session *mgo.Session
	var err error
	if session, err = GetSession(b.dialInfo, wmajor, b.ssl, b.sslCAFile, b.sslPEMKeyFile); err != nil {
		return
	}
	defer session.Close()
	c := session.DB(SimDBName).C(CollectionName)

	for run := 0; run < duration; run++ {
		// be a minute transactions
		stage := "setup"
		if run == (duration - 1) {
			stage = "teardown"
			isTeardown = true
			totalTPS = b.tps
		} else if run > 0 && run < (duration-1) {
			stage = "thrashing"
			totalTPS = b.tps
		} else {
			totalTPS = b.tps / 2
		}

		batchCount := 0
		totalCount := 0
		beginTime := time.Now()
		counter := 0
		for time.Now().Sub(beginTime) < time.Minute {
			innerTime := time.Now()
			txCount := 0
			for time.Now().Sub(innerTime) < time.Second && txCount < totalTPS {
				doc := simDocs[batchCount%len(simDocs)]
				batchCount++
				if isTeardown {
					c.RemoveAll(bson.M{"_search": doc["_search"]})
				} else if len(transactions) > 0 { // --file and --tx
					txCount += execTXByTemplateAndTX(c, cloneDoc(doc), transactions)
				} else if len(transactions) == 0 { // --file
					txCount += execTXByTemplate(c, cloneDoc(doc))
				} else if b.filename == "" {
					txCount += execTXForDemo(c, cloneDoc(doc))
				}
				// time.Sleep(1 * time.Millisecond)
			} // for time.Now().Sub(innerTime) < time.Second && txCount < totalTPS
			totalCount += txCount
			counter++
			seconds := 1 - time.Now().Sub(innerTime).Seconds()
			if seconds > 0 {
				time.Sleep(time.Duration(seconds) * time.Second)
			}
		} // for time.Now().Sub(beginTime) < time.Minute

		if b.verbose {
			log.Println("=>", time.Now().Sub(beginTime), time.Now().Sub(beginTime) > time.Minute,
				totalCount, totalCount/counter < totalTPS, counter)
		}
		tenPctOff := float64(totalTPS) * .95
		if b.verbose || totalCount/counter < int(tenPctOff) {
			log.Printf("%s average TPS was %d, lower than original %d\n", stage, totalCount/counter, totalTPS)
		}

		seconds := 60 - time.Now().Sub(beginTime).Seconds()
		if seconds > 0 {
			time.Sleep(time.Duration(seconds) * time.Second)
		}
		if b.verbose {
			log.Println("=>", time.Now().Sub(beginTime))
		}
	} //for run := 0; run < duration; run++

	c.DropCollection()
}

// cloneDoc clones a doc and assign a _id
func cloneDoc(doc bson.M) bson.M {
	_id := bson.NewObjectId()
	var ndoc = make(bson.M)
	bytes, _ := json.Marshal(doc)
	json.Unmarshal(bytes, &ndoc)
	ndoc["_id"] = _id
	return ndoc
}

// CreateIndexes creates indexes
func (b Base) CreateIndexes(docs []bson.M) error {
	var session *mgo.Session
	var err error
	if session, _ = GetSession(b.dialInfo, false, b.ssl, b.sslCAFile, b.sslPEMKeyFile); err != nil {
		return err
	}
	defer session.Close()
	c := session.DB(SimDBName).C(CollectionName)

	if len(docs) == 0 {
		c.EnsureIndexKey("favoriteCity")
	}
	c.EnsureIndexKey("_search")

	for _, doc := range docs {
		keys := []string{}
		for field := range doc {
			keys = append(keys, field)
		}

		err = c.EnsureIndex(mgo.Index{Key: keys})
		if err != nil {
			return err
		}
	}

	return err
}
