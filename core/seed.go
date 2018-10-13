// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// SeedBase -
type SeedBase struct {
	File       string
	Collection string
	Total      int
	IsDrop     bool
	DBName     string
	Verbose    bool
}

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

// NewSeedBase establish seeding parameters
func NewSeedBase(file string, collection string, total int, isDrop bool, dbName string, verbose bool) SeedBase {
	sb := SeedBase{file, collection, total, isDrop, dbName, verbose}
	return sb
}

// SeedData seeds demo data
func (sb SeedBase) SeedData(session *mgo.Session) error {
	if sb.File == "" {
		sb.seed(session)
	} else {
		if sb.Collection == "" {
			return errors.New("usage: keyhole --uri connection_uri --seed [--file filename --collection collection_name]")
		}
		sb.seedFromTemplate(session)
	}

	return nil
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
func (sb SeedBase) seed(session *mgo.Session) {
	session.SetMode(mgo.Primary, true)

	c := session.DB(sb.DBName).C("lookups")
	if sb.IsDrop {
		c.DropCollection()
	}

	for i := 0; i < 10; i++ {
		c.Insert(bson.M{"_id": i + 1000, "type": "sports", "name": Favorites.Sports[i]})
		c.Insert(bson.M{"_id": i + 1100, "type": "book", "name": Favorites.Books[i]})
		c.Insert(bson.M{"_id": i + 1200, "type": "movie", "name": Favorites.Movies[i]})
		c.Insert(bson.M{"_id": i + 1300, "type": "city", "name": Favorites.Cities[i]})
		c.Insert(bson.M{"_id": i + 1400, "type": "music", "name": Favorites.Music[i]})
	}

	modelsCollection := session.DB(sb.DBName).C("models")
	robotsCollection := session.DB(sb.DBName).C("robots")
	if sb.IsDrop {
		modelsCollection.DropCollection()
		robotsCollection.DropCollection()
	}

	for i := 1000; i < 1010; i++ {
		model := "model-" + fmt.Sprintf("%x", (rand.Intn(5000)+5000)*i)
		name := fmt.Sprintf("Robo %d-%x", i, rand.Intn(1000000))
		descr := fmt.Sprintf("%s %s", model, name)
		year := time.Now().Year() - rand.Intn(5)
		err := modelsCollection.Insert(&Model{model, name, descr, year})
		if err != nil {
			log.Fatal(err)
		}

		for r := 0; r < 2+rand.Intn(20); r++ {
			id := "robot-" + fmt.Sprintf("%x", (rand.Intn(5000)+5000)*r)
			notes := fmt.Sprintf("%s %s", id, model)
			pct := rand.Float32()
			tasks := []Task{{"Business", 10 + rand.Intn(60)}, {"Home", 10 + rand.Intn(60)}}
			err := robotsCollection.Insert(&Robot{id, model, notes, pct, tasks})
			if err != nil {
				continue
			}
		}

		if sb.Verbose == true {
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

	var contentArray []interface{}
	numbersCollection := session.DB(sb.DBName).C("numbers")
	if sb.IsDrop {
		numbersCollection.DropCollection()
	}
	bulk := numbersCollection.Bulk()
	for n := 0; n < 1000; n++ {
		contentArray = append(contentArray, bson.M{"a": rand.Intn(100), "b": rand.Intn(50), "c": rand.Intn(1000)})
	}
	bulk.Insert(contentArray...)
	_, err := bulk.Run()
	if err != nil {
		log.Println(err)
		return
	}
	numbersCount, _ := numbersCollection.Count()
	fmt.Printf("Seeded models: %d, robots: %d, numbers: %d\n", modelsCount, robotsCount, numbersCount)

	carsCollection := session.DB(sb.DBName).C("cars")
	favoritesCollection := session.DB(sb.DBName).C("favorites")
	if sb.IsDrop {
		carsCollection.DropCollection()
		favoritesCollection.DropCollection()
	}

	carsCount := sb.seedCollection(carsCollection, sb.Total, 1)
	fmt.Printf("Seeded cars: %d\n", carsCount)
	favoritesCount := sb.seedCollection(favoritesCollection, sb.Total, 2)
	fmt.Printf("Seeded favorites: %d\n", favoritesCount)
}

var isNew = []bool{true, false}
var styles = []string{"Sedan", "Coupe", "Convertible", "Minivan", "SUV", "Truck"}
var colors = []string{"Beige", "Black", "Blue", "Brown", "Gold", "Gray", "Green", "Orange", "Pink", "Purple", "Red", "Silver", "White", "Yellow"}

func getCar() bson.M {
	return bson.M{
		"isNew": isNew[rand.Intn(len(isNew))],
		"style": styles[rand.Intn(len(styles))],
		"color": colors[rand.Intn(len(colors))],
	}
}

func (sb SeedBase) seedCollection(c *mgo.Collection, total int, fnum int) int {
	bsize := 100
	remaining := total
	for i := 0; i < total; {
		bulk := c.Bulk()
		num := bsize
		if remaining < bsize {
			num = remaining
		}
		var contentArray []interface{}
		for n := 0; n < num; n++ {
			if fnum == 1 {
				contentArray = append(contentArray, getCar())
			} else if fnum == 2 {
				contentArray = append(contentArray, GetDemoDoc())
			}
			i++
			remaining--
		}
		bulk.Insert(contentArray...)
		_, err := bulk.Run()
		if err != nil {
			log.Println(err)
			return 0
		}
		fmt.Fprintf(os.Stderr, "\r%3.1f%% ", float64(100*i)/float64(total))
	}
	fmt.Fprintf(os.Stderr, "\r100%%\r     \r")
	cnt, _ := c.Count()
	return cnt
}

// SeedFromTemplate seeds data from a template in a file
func (sb SeedBase) seedFromTemplate(session *mgo.Session) {
	sdoc := GetDocByTemplate(sb.File, true)
	bytes, _ := json.MarshalIndent(sdoc, "", "   ")
	if sb.Verbose {
		fmt.Println(string(bytes))
	}
	doc := make(map[string]interface{})
	json.Unmarshal(bytes, &doc)
	session.SetMode(mgo.Primary, true)
	collName := sb.Collection
	if collName == "" {
		collName = "examples"
	}
	fmt.Println("Seed data to collection", collName)
	examplesCollection := session.DB(sb.DBName).C(collName)
	if sb.IsDrop {
		examplesCollection.DropCollection()
	}
	bsize := 100
	remaining := sb.Total
	for i := 0; i < sb.Total; {
		bulk := examplesCollection.Bulk()
		num := bsize
		if remaining < bsize {
			num = remaining
		}
		var contentArray []interface{}
		for n := 0; n < num; n++ {
			ndoc := make(map[string]interface{})
			RandomizeDocument(&ndoc, doc, false)
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
		fmt.Fprintf(os.Stderr, "\r%3.1f%% ", float64(100*i)/float64(sb.Total))
	}

	fmt.Fprintf(os.Stderr, "\r100%%   \n")
	examplesCount, _ := examplesCollection.Count()
	fmt.Printf("\rSeeded examples: %d, total count: %d\n", sb.Total, examplesCount)
}
