// Copyright 2018 Kuei-chun Chen. All rights reserved.

package sim

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"sort"
	"strconv"
	"time"

	"github.com/simagix/keyhole/mdb"
	"github.com/simagix/keyhole/sim/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var simDocs []bson.M

// initialize an array of documents for simulation test.  If a template is available
// read the sample json and replace them with random values.  Otherwise, use the demo
// example.
func (rn *Runner) initSimDocs() {
	var err error
	var sdoc bson.M

	if rn.verbose {
		log.Println("initSimDocs")
	}
	rand.Seed(time.Now().Unix())
	total := 512
	if rn.filename == "" {
		for len(simDocs) < total {
			simDocs = append(simDocs, util.GetDemoDoc())
		}
		return
	}

	if sdoc, err = util.GetDocByTemplate(rn.filename, true); err != nil {
		return
	}
	bytes, _ := json.Marshal(sdoc)
	if rn.verbose {
		log.Println(string(bytes))
	}
	doc := make(map[string]interface{})
	json.Unmarshal(bytes, &doc)

	for len(simDocs) < total {
		ndoc := make(map[string]interface{})
		util.RandomizeDocument(&ndoc, doc, false)
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
func PopulateData(uri string, sslCAFile string, sslPEMKeyFile string) error {
	var err error
	var client *mongo.Client
	if client, err = mdb.NewMongoClient(uri, sslCAFile, sslPEMKeyFile); err != nil {
		return err
	}
	c := client.Database(SimDBName).Collection(CollectionName)
	btime := time.Now()
	for time.Now().Sub(btime) < time.Minute {
		var contentArray []interface{}
		docidx := 0
		for i := 0; i < 1000; i++ {
			contentArray = append(contentArray, simDocs[docidx%len(simDocs)])
			docidx++
		}
		if _, err = c.InsertMany(context.Background(), contentArray); err != nil {
			return err
		}
	}

	return nil
}

// Simulate simulates CRUD for load tests
func (rn *Runner) Simulate(duration int, transactions []Transaction, thread int) error {
	var err error
	var client *mongo.Client
	var ctx = context.Background()
	var totalTPS int

	if client, err = mdb.NewMongoClient(rn.uri, rn.sslCAFile, rn.sslPEMKeyFile); err != nil {
		return err
	}
	defer client.Disconnect(ctx)
	c := client.Database(SimDBName).Collection(CollectionName)

	for run := 0; run < duration; run++ {
		// be a minute transactions
		stage := "setup"
		if run == (duration - 1) {
			stage = "teardown"
			totalTPS = rn.tps
		} else if run > 0 && run < (duration-1) {
			stage = "thrashing"
			totalTPS = rn.tps
		} else {
			totalTPS = rn.tps / 2
		}
		if thread == 0 {
			log.Println(stage, "TPS:", totalTPS)
		}
		batchCount := 0
		totalCount := 0
		beginTime := time.Now()
		counter := 0
		results := []bson.M{}
		for time.Now().Sub(beginTime) < time.Minute {
			innerTime := time.Now()
			txCount := 0
			for time.Now().Sub(innerTime) < time.Second && txCount < totalTPS {
				doc := simDocs[batchCount%len(simDocs)]
				batchCount++
				if stage == "setup" {
					t := time.Now()
					if _, err = c.InsertOne(ctx, util.CloneDoc(doc)); err != nil {
						break
					}
					if thread == 0 {
						results = append(results, bson.M{"InsertOne": time.Now().Sub(t)})
					}
					txCount++
				} else if stage == "thrashing" {
					var res bson.M
					if len(transactions) > 0 {
						if res, err = execTXByTemplateAndTX(c, util.CloneDoc(doc), transactions); err != nil {
							break
						}
						txCount += res["total"].(int)
					} else {
						if res, err = execTx(c, util.CloneDoc(doc)); err != nil {
							break
						}
						txCount += res["total"].(int)
					}

					if thread == 0 {
						delete(res, "total")
						results = append(results, res)
					}
				} else if stage == "teardown" {
					c.DeleteMany(ctx, bson.M{"_search": strconv.FormatInt(rand.Int63(), 16)})
				}
				time.Sleep(time.Millisecond)
			} // for time.Now().Sub(innerTime) < time.Second && txCount < totalTPS
			totalCount += txCount
			counter++
			milli := 1000 - time.Now().Sub(innerTime).Milliseconds()
			if milli > 0 {
				time.Sleep(time.Duration(milli) * time.Millisecond)
			}
		} // for time.Now().Sub(beginTime) < time.Minute

		if thread == 0 && len(results) > 0 {
			length := 0
			durations := map[string][]time.Duration{}
			for _, res := range results {
				for k, v := range res {
					length++
					durations[k] = append(durations[k], v.(time.Duration))
				}
			}
			log.Println("Executions Time (including network latency) from thread 0: ", length, "samples:")
			tm := time.Now()
			client.Ping(ctx, nil)
			log.Printf("\t[%12s] %v\n", "Ping", time.Now().Sub(tm))
			keys := make([]string, 0, len(durations))
			for k := range durations {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				v := durations[k]
				sort.Slice(v, func(i int, j int) bool {
					return v[i] < v[j]
				})
				var sum time.Duration
				for _, t := range v {
					sum += t
				}
				length = len(v)
				p95 := int64(float64(length) * .95)
				p99 := int64(float64(length) * .99)
				log.Printf("\t[%12s] (min,avg,p95,p99,max) = (%v,%v,%v,%v,%v)\n",
					k, v[0], sum/time.Duration(length), v[p95], v[p99], v[length-1])
			}
		}

		if rn.verbose {
			log.Println("=>", time.Now().Sub(beginTime), time.Now().Sub(beginTime) > time.Minute,
				totalCount, totalCount/counter < totalTPS, counter)
		}
		tenPctOff := float64(totalTPS) * .95
		if rn.verbose && totalCount/counter < int(tenPctOff) {
			log.Printf("%s average TPS was %d, lower than original %d\n", stage, totalCount/counter, totalTPS)
		}

		seconds := 60 - time.Now().Sub(beginTime).Seconds()
		if seconds > 0 {
			time.Sleep(time.Duration(seconds) * time.Second)
		}
		if rn.verbose {
			log.Println("=>", time.Now().Sub(beginTime))
		}
	} //for run := 0; run < duration; run++

	c.Drop(ctx)
	return nil
}

func printAverageTimeDuration(results []bson.M) {
}
