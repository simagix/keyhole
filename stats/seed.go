// Copyright 2018 Kuei-chun Chen. All rights reserved.

package stats

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/simagix/keyhole/utils"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Model - robot model
type Model struct {
	ID          string `json:"_id" bson:"_id"`
	Name        string
	Description string
	Year        int
}

// Task - robot task
type Task struct {
	For         string `json:"for" bson:"for"`
	MinutesUsed int    `json:"minutesUsed" bson:"minutesUsed"`
}

// Robot -
type Robot struct {
	ID         string  `json:"_id" bson:"_id"`
	ModelID    string  `json:"modelId,omitempty" bson:"modelId,omitempty"`
	Notes      string  `json:"notes" bson:"notes"`
	BatteryPct float32 `json:"batteryPct,omitempty" bson:"batteryPct,omitempty"`
	Tasks      []Task  `json:"tasks" bson:"tasks"`
}

// GetDocByTemplate returns a bson.M document
func GetDocByTemplate(filename string, meta bool) bson.M {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	var f interface{}
	err = json.Unmarshal(bytes, &f)
	if err != nil {
		fmt.Println("Error parsing JSON: ", err)
		panic(err)
	}
	doc := make(map[string]interface{})
	utils.RandomizeDocument(&doc, f, meta)
	return doc
}

// Seed - seed data for demo
//  models: {
//    "_id": string
//   "name": string,
//   "description": string
//   "year": integer
// }
//
// robots: {
//   "_id": string
//   "modelId": string
//   "notes": string,
//   "batteryPct": float,
//   "tasks": [{"for": string, "minutesUsed": integer}]
// }
func Seed(session *mgo.Session, isDrop bool, dbName string, verbose bool) {
	session.SetMode(mgo.Primary, true)
	modelsCollection := session.DB(dbName).C("models")
	robotsCollection := session.DB(dbName).C("robots")
	if isDrop {
		modelsCollection.DropCollection()
		robotsCollection.DropCollection()
	}

	for i := 1000; i < 1050; i++ {
		model := "model-" + fmt.Sprintf("%x", (rand.Intn(5000)+5000)*i)
		name := fmt.Sprintf("Robo %d-%x", i, rand.Intn(1000000))
		descr := fmt.Sprintf("%s %s", model, name)
		year := time.Now().Year() - rand.Intn(5)
		err := modelsCollection.Insert(&Model{model, name, descr, year})
		if err != nil {
			log.Fatal(err)
		}

		for r := 0; r < 20+rand.Intn(30); r++ {
			id := "robot-" + fmt.Sprintf("%x", (rand.Intn(5000)+5000)*r)
			notes := fmt.Sprintf("%s %s", id, model)
			pct := rand.Float32()
			tasks := []Task{{"Business", 10 + rand.Intn(60)}, {"Home", 10 + rand.Intn(60)}}
			err := robotsCollection.Insert(&Robot{id, model, notes, pct, tasks})
			if err != nil {
				continue
			}
		}

		if verbose == true {
			modelRes := Model{}
			robotRes := []Robot{}
			_ = modelsCollection.Find(bson.M{"_id": model}).One(&modelRes)
			fmt.Println(modelRes)
			_ = robotsCollection.Find(bson.M{"modelId": model}).All(&robotRes)
			fmt.Println(robotRes)
		}
	}
	modelsCount, _ := modelsCollection.Count()
	robotsCount, _ := robotsCollection.Count()
	fmt.Printf("Seeded models: %d, robots: %d\n", modelsCount, robotsCount)

	isNew := []bool{true, false}
	styles := []string{"Sedan", "Coupe", "Convertible", "Minivan", "SUV", "Truck"}
	colors := []string{"Beige", "Black", "Blue", "Brown", "Gold", "Gray", "Green", "Orange", "Pink", "Purple", "Red", "Silver", "White", "Yellow"}

	carsCollection := session.DB(dbName).C("cars")
	keyholeCollection := session.DB(dbName).C("keyhole")
	if isDrop {
		carsCollection.DropCollection()
		keyholeCollection.DropCollection()
	}
	for i := 0; i < 250; i++ {
		keyholeCollection.Insert(GetRandomDoc())
		bulk := carsCollection.Bulk()
		var contentArray []interface{}
		for n := 0; n < 1000; n++ {
			contentArray = append(contentArray,
				bson.M{
					"isNew": isNew[rand.Intn(len(isNew))],
					"style": styles[rand.Intn(len(styles))],
					"color": colors[rand.Intn(len(colors))],
				})
		}

		bulk.Insert(contentArray...)
		_, err := bulk.Run()
		if err != nil {
			log.Println(err)
			return
		}
	}
	var contentArray []interface{}
	numbersCollection := session.DB(dbName).C("numbers")
	if isDrop {
		numbersCollection.DropCollection()
	}
	bulk := numbersCollection.Bulk()
	for n := 0; n < 100000; n++ {
		contentArray = append(contentArray, bson.M{"a": rand.Intn(100), "b": rand.Intn(50), "c": rand.Intn(1000)})
	}
	bulk.Insert(contentArray...)
	_, err := bulk.Run()
	if err != nil {
		log.Println(err)
		return
	}
	carsCount, _ := carsCollection.Count()
	keyholeCount, _ := keyholeCollection.Count()
	numbersCount, _ := numbersCollection.Count()
	fmt.Printf("Seeded cars: %d, keyhole: %d, numbers: %d\n", carsCount, keyholeCount, numbersCount)
}

// SeedFromTemplate seeds data from a template in a file
func SeedFromTemplate(session *mgo.Session, filename string, total int, isDrop bool, dbName string, verbose bool) {
	sdoc := GetDocByTemplate(filename, true)
	bytes, _ := json.MarshalIndent(sdoc, "", "   ")
	if verbose {
		fmt.Println(string(bytes))
	}
	doc := make(map[string]interface{})
	json.Unmarshal(bytes, &doc)
	session.SetMode(mgo.Primary, true)
	examplesCollection := session.DB(dbName).C("examples")
	if isDrop {
		examplesCollection.DropCollection()
	}
	bsize := 100
	remaining := total
	for i := 0; i < total; {
		bulk := examplesCollection.Bulk()
		num := bsize
		if remaining < bsize {
			num = remaining
		}
		var contentArray []interface{}
		for n := 0; n < num; n++ {
			ndoc := make(map[string]interface{})
			utils.RandomizeDocument(&ndoc, doc, false)
			delete(ndoc, "_id")
			contentArray = append(contentArray, ndoc)
			i++
			remaining--
		}
		bulk.Insert(contentArray...)
		_, err := bulk.Run()
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Fprintf(os.Stderr, "\r%3.1f%% ", float64(100*i)/float64(total))
	}

	fmt.Fprintf(os.Stderr, "\r100%%   \n")
	examplesCount, _ := examplesCollection.Count()
	fmt.Printf("\rSeeded examples: %d, total count: %d\n", total, examplesCount)
}
