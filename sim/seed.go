// Copyright 2018 Kuei-chun Chen. All rights reserved.

package sim

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"
	"github.com/mongodb/mongo-go-driver/x/bsonx"
	"github.com/simagix/keyhole/sim/util"
)

// SeedBase -
type SeedBase struct {
	File       string
	Collection string
	Total      int
	IsDrop     bool
	DBName     string
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
func NewSeedBase(file string, collection string, total int, isDrop bool, dbName string) SeedBase {
	sb := SeedBase{file, collection, total, isDrop, dbName}
	return sb
}

// SeedData seeds demo data
func (sb SeedBase) SeedData(client *mongo.Client) error {
	if sb.File == "" {
		sb.seed(client)
	} else {
		if sb.Collection == "" {
			return errors.New("usage: keyhole --uri connection_uri --seed [--file filename --collection collection_name]")
		}
		sb.seedFromTemplate(client)
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
func (sb SeedBase) seed(client *mongo.Client) {
	var err error
	ctx := context.Background()
	c := client.Database(sb.DBName).Collection("lookups")
	if sb.IsDrop {
		c.Drop(ctx)
	}

	for i := 0; i < 10; i++ {
		c.InsertOne(ctx, bson.M{"_id": i + 1000, "type": "sports", "name": util.Favorites.Sports[i]})
		c.InsertOne(ctx, bson.M{"_id": i + 1100, "type": "book", "name": util.Favorites.Books[i]})
		c.InsertOne(ctx, bson.M{"_id": i + 1200, "type": "movie", "name": util.Favorites.Movies[i]})
		c.InsertOne(ctx, bson.M{"_id": i + 1300, "type": "city", "name": util.Favorites.Cities[i]})
		c.InsertOne(ctx, bson.M{"_id": i + 1400, "type": "music", "name": util.Favorites.Music[i]})
	}

	modelsCollection := client.Database(sb.DBName).Collection("models")
	robotsCollection := client.Database(sb.DBName).Collection("robots")
	if sb.IsDrop {
		modelsCollection.Drop(ctx)
		robotsCollection.Drop(ctx)
	}

	for i := 1000; i < 1010; i++ {
		model := "model-" + fmt.Sprintf("%x", (rand.Intn(5000)+5000)*i)
		name := fmt.Sprintf("Robo %d-%x", i, rand.Intn(1000000))
		descr := fmt.Sprintf("%s %s", model, name)
		year := time.Now().Year() - rand.Intn(5)
		if _, err = modelsCollection.InsertOne(ctx, &Model{model, name, descr, year}); err != nil {
			log.Fatal(err)
		}

		for r := 0; r < 2+rand.Intn(20); r++ {
			id := "robot-" + fmt.Sprintf("%x", (rand.Intn(5000)+5000)*r)
			notes := fmt.Sprintf("%s %s", id, model)
			pct := rand.Float32()
			tasks := []Task{{"Business", 10 + rand.Intn(60)}, {"Home", 10 + rand.Intn(60)}}
			if _, err = robotsCollection.InsertOne(ctx, &Robot{id, model, notes, pct, tasks}); err != nil {
				continue
			}
		}
	}
	modelsCount, _ := modelsCollection.Count(ctx, bson.M{})
	robotsCount, _ := robotsCollection.Count(ctx, bson.M{})

	numbersCollection := client.Database(sb.DBName).Collection("numbers")
	if sb.IsDrop {
		numbersCollection.Drop(ctx)
	}

	var docs []interface{}
	for n := 0; n < 1000; n++ {
		docs = append(docs, bson.M{"a": rand.Intn(100), "b": rand.Intn(100), "c": rand.Intn(100)})
	}
	if _, err = numbersCollection.InsertMany(ctx, docs); err != nil {
		log.Fatal(err)
	}
	// create index example
	indexView := numbersCollection.Indexes()
	idx := mongo.IndexModel{
		Keys: bsonx.Doc{{Key: "a", Value: bsonx.Int32(1)}},
	}
	indexView.CreateOne(ctx, idx)
	idx = mongo.IndexModel{
		Keys: bsonx.Doc{{Key: "a", Value: bsonx.Int32(1)}, {Key: "b", Value: bsonx.Int32(1)}},
	}
	indexView.CreateOne(ctx, idx)
	idx = mongo.IndexModel{
		Keys: bsonx.Doc{{Key: "a", Value: bsonx.Int32(1)}, {Key: "b", Value: bsonx.Int32(-1)}},
	}
	indexView.CreateOne(ctx, idx)
	idx = mongo.IndexModel{
		Keys: bsonx.Doc{{Key: "a", Value: bsonx.Int32(1)}, {Key: "b", Value: bsonx.Int32(1)}, {Key: "c", Value: bsonx.Int32(1)}},
	}
	indexView.CreateOne(ctx, idx)

	numbersCount, _ := numbersCollection.Count(ctx, bson.M{})
	fmt.Printf("Seeded models: %d, robots: %d, numbers: %d\n", modelsCount, robotsCount, numbersCount)

	carsCollection := client.Database(sb.DBName).Collection("cars")
	dealersCollection := client.Database(sb.DBName).Collection("dealers")
	favoritesCollection := client.Database(sb.DBName).Collection("favorites")
	if sb.IsDrop {
		carsCollection.Drop(ctx)
		dealersCollection.Drop(ctx)
		favoritesCollection.Drop(ctx)
	}

	// Upsert examples
	for i := 0; i < len(dealers); i++ {
		dealerID := fmt.Sprintf("DEALER-%d", 1+i)
		opts := options.Update()
		opts.SetUpsert(true)
		if _, err := dealersCollection.UpdateOne(ctx, bson.M{"_id": dealerID}, bson.M{"$set": bson.M{"name": dealers[i]}}, opts); err != nil {
			log.Fatal(err)
		}
	}

	// create index example
	indexView = carsCollection.Indexes()
	idx = mongo.IndexModel{
		Keys: bsonx.Doc{{Key: "filters.k", Value: bsonx.Int32(1)}, {Key: "filters.v", Value: bsonx.Int32(1)}},
	}
	indexView.CreateOne(ctx, idx)

	dealersCount, _ := dealersCollection.Count(ctx, bson.M{})
	carsCount := sb.seedCollection(carsCollection, 1)
	fmt.Printf("Seeded cars: %d, dealers: %d\n", carsCount, dealersCount)
	favoritesCount := sb.seedCollection(favoritesCollection, 2)
	fmt.Printf("Seeded favorites: %d\n", favoritesCount)
}

var dealers = []string{"Atlanta Auto", "Buckhead Auto", "Johns Creek Auto"}
var brands = []string{"Audi", "BMW", "Chevrolet", "Ford", "Honda", "Mercedes-Benz", "Nissan", "Porsche", "Toyota", "Volkswagen"}
var styles = []string{"Sedan", "Coupe", "Convertible", "Minivan", "SUV", "Truck"}
var colors = []string{"Beige", "Black", "Blue", "Brown", "Gold", "Gray", "Green", "Orange", "Pink", "Purple", "Red", "Silver", "White", "Yellow"}

func getVehicle() bson.M {
	curYear := time.Now().Year()
	delta := rand.Intn(8)
	year := curYear - delta
	used := true
	if delta == 0 {
		used = false
	}
	brand := brands[rand.Intn(len(styles))]
	color := colors[rand.Intn(len(colors))]
	style := styles[rand.Intn(len(styles))]

	return bson.M{
		"dealer": fmt.Sprintf("DEALER-%d", 1+rand.Intn(len(dealers))),
		"brand":  brand,
		"color":  color,
		"style":  style,
		"year":   year,
		"used":   used,
		"filters": []bson.M{
			bson.M{"k": "brand", "v": brand},
			bson.M{"k": "color", "v": color},
			bson.M{"k": "style", "v": style},
			bson.M{"k": "year", "v": year},
			bson.M{"k": "used", "v": used}},
	}
}

func (sb SeedBase) seedCollection(c *mongo.Collection, fnum int) int {
	var err error
	var ctx = context.Background()
	var bsize = 1000
	var remaining = sb.Total

	for i := 0; i < sb.Total; {
		num := bsize
		if remaining < bsize {
			num = remaining
		}
		var contentArray []interface{}
		for n := 0; n < num; n++ {
			if fnum == 1 {
				contentArray = append(contentArray, getVehicle())
			} else if fnum == 2 {
				contentArray = append(contentArray, util.GetDemoDoc())
			}
			i++
			remaining--
		}
		if _, err = c.InsertMany(ctx, contentArray); err != nil {
			panic(err)
		}
		fmt.Fprintf(os.Stderr, "\r%3.1f%% ", float64(100*i)/float64(sb.Total))
	}
	fmt.Fprintf(os.Stderr, "\r100%%\r     \r")
	cnt, _ := c.Count(ctx, bson.M{})
	return int(cnt)
}

// SeedFromTemplate seeds data from a template in a file
func (sb SeedBase) seedFromTemplate(client *mongo.Client) {
	var err error
	var ctx = context.Background()
	var bsize = 1000
	var remaining = sb.Total
	var sdoc bson.M
	if sdoc, err = util.GetDocByTemplate(sb.File, true); err != nil {
		return
	}
	bytes, _ := json.MarshalIndent(sdoc, "", "   ")
	doc := make(map[string]interface{})
	json.Unmarshal(bytes, &doc)
	collName := sb.Collection
	if collName == "" {
		collName = "examples"
	}
	log.Println("Seed data to collection", collName)
	c := client.Database(sb.DBName).Collection(collName)
	if sb.IsDrop {
		c.Drop(ctx)
	}

	for i := 0; i < sb.Total; {
		num := bsize
		if remaining < bsize {
			num = remaining
		}
		var contentArray []interface{}
		for n := 0; n < num; n++ {
			ndoc := make(map[string]interface{})
			util.RandomizeDocument(&ndoc, doc, false)
			delete(ndoc, "_id")
			contentArray = append(contentArray, ndoc)
			i++
			remaining--
		}
		if _, err = c.InsertMany(ctx, contentArray); err != nil {
			panic(err)
		}
		fmt.Fprintf(os.Stderr, "\r%3.1f%% ", float64(100*i)/float64(sb.Total))
	}

	fmt.Fprintf(os.Stderr, "\r100%%   \n")
	cnt, _ := c.Count(ctx, bson.M{})
	fmt.Printf("\rSeeded %s: %d, total count: %d\n", collName, sb.Total, cnt)
}
