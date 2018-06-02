// Copyright 2018 Kuei-chun Chen. All rights reserved.

package stats

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const keyholeDB = "_KEYHOLE_"
const x = 1024

// DBName -
var DBName = fmt.Sprintf("_KEYHOLE_%X", x+x*rand.Intn(x))

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
func Seed(session *mgo.Session, verbose bool) {
	session.SetMode(mgo.Primary, true)
	session.DB(keyholeDB).DropDatabase()
	modelsCollection := session.DB(keyholeDB).C("models")
	robotsCollection := session.DB(keyholeDB).C("robots")

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

	carsCollection := session.DB(keyholeDB).C("cars")
	keyholeCollection := session.DB(keyholeDB).C("keyhole")
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
	numbersCollection := session.DB(keyholeDB).C("numbers")
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
