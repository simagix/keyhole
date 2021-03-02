// Copyright 2021 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/simagix/gox"
	"github.com/simagix/keyhole/mdb"
	"github.com/simagix/keyhole/sim/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Seed seeds feeder
type Seed struct {
	collection   string
	conns        int
	database     string
	file         string
	isDrop       bool
	showProgress bool
	total        int
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

// NewSeed establish seeding parameters
func NewSeed() *Seed {
	return &Seed{conns: runtime.NumCPU(), isDrop: false, total: 1000, showProgress: true}
}

// SetCollection set collection
func (f *Seed) SetCollection(collection string) {
	f.collection = collection
}

// SetNumberConnections set conns
func (f *Seed) SetNumberConnections(conns int) {
	if conns == 0 {
		return
	}
	f.conns = conns
}

// SetDatabase set database
func (f *Seed) SetDatabase(database string) {
	f.database = database
}

// SetFile set file
func (f *Seed) SetFile(file string) {
	f.file = file
}

// SetIsDrop set isDrop
func (f *Seed) SetIsDrop(isDrop bool) {
	f.isDrop = isDrop
}

// SetShowProgress set showProgress
func (f *Seed) SetShowProgress(showProgress bool) {
	f.showProgress = showProgress
}

// SetTotal set total
func (f *Seed) SetTotal(total int) {
	f.total = total
}

// SeedData seeds all demo data
func (f *Seed) SeedData(client *mongo.Client) error {
	if f.file == "" {
		return f.SeedAllDemoData(client)
	} else if f.collection == "" {
		return errors.New("usage: keyhole --uri connection_uri --seed [--file filename --collection collection_name]")
	}
	return f.seedFromTemplate(client)
}

// SeedAllDemoData - seed data for demo
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
func (f *Seed) SeedAllDemoData(client *mongo.Client) error {
	var err error
	if err = f.SeedFavorites(client); err != nil {
		return err
	}
	if err = f.seedRobots(client); err != nil {
		return err
	}
	if err = f.seedNumbers(client); err != nil {
		return err
	}
	if err = f.SeedVehicles(client); err != nil {
		return err
	}
	return err
}

// SeedFavorites seeds demo data of collection favorites
func (f *Seed) SeedFavorites(client *mongo.Client) error {
	var err error
	var ctx = context.Background()
	c := client.Database(f.database).Collection("lookups")
	favoritesCollection := client.Database(f.database).Collection("favorites")
	if f.isDrop {
		if err = c.Drop(ctx); err != nil {
			return err
		}
		if err = favoritesCollection.Drop(ctx); err != nil {
			return err
		}
	}

	for i := 0; i < 10; i++ {
		c.InsertOne(ctx, bson.M{"_id": i + 1000, "type": "sports", "name": util.Favorites.Sports[i]})
		c.InsertOne(ctx, bson.M{"_id": i + 1100, "type": "book", "name": util.Favorites.Books[i]})
		c.InsertOne(ctx, bson.M{"_id": i + 1200, "type": "movie", "name": util.Favorites.Movies[i]})
		c.InsertOne(ctx, bson.M{"_id": i + 1300, "type": "city", "name": util.Favorites.Cities[i]})
		c.InsertOne(ctx, bson.M{"_id": i + 1400, "type": "music", "name": util.Favorites.Music[i]})
	}
	favoritesCount := f.seedCollection(favoritesCollection, 2)
	fmt.Printf("Seeded favorites: %d\n", favoritesCount)
	return err
}

func (f *Seed) seedRobots(client *mongo.Client) error {
	var err error
	var ctx = context.Background()
	modelsCollection := client.Database(f.database).Collection("models")
	robotsCollection := client.Database(f.database).Collection("robots")
	if f.isDrop {
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
	modelsCount, _ := modelsCollection.CountDocuments(ctx, bson.M{})
	robotsCount, _ := robotsCollection.CountDocuments(ctx, bson.M{})
	fmt.Printf("Seeded models: %d, robots: %d\n", modelsCount, robotsCount)
	return err
}

func (f *Seed) seedNumbers(client *mongo.Client) error {
	var err error
	var ctx = context.Background()
	numbersCollection := client.Database(f.database).Collection("numbers")
	if f.isDrop {
		numbersCollection.Drop(ctx)
	}

	var docs []interface{}
	docs = append(docs, bson.M{"a": rand.Intn(100), "b": primitive.NewDecimal128(100, 0), "c": math.NaN()})
	for n := 1; n < 1000; n++ {
		docs = append(docs, bson.M{"a": rand.Intn(100), "b": rand.Intn(100), "c": rand.Intn(100)})
	}
	if _, err = numbersCollection.InsertMany(ctx, docs); err != nil {
		fmt.Println("Seeded numbers: 0", err) // could be < v3.4
		return nil
	}
	// create index example
	indexView := numbersCollection.Indexes()
	idx := mongo.IndexModel{
		Keys: bson.D{{Key: "a", Value: 1}},
	}
	indexView.CreateOne(ctx, idx)
	idx = mongo.IndexModel{
		Keys: bson.D{{Key: "a", Value: 1}, {Key: "b", Value: 1}},
	}
	indexView.CreateOne(ctx, idx)
	idx = mongo.IndexModel{
		Keys: bson.D{{Key: "a", Value: 1}, {Key: "b", Value: -1}},
	}
	indexView.CreateOne(ctx, idx)
	idx = mongo.IndexModel{
		Keys: bson.D{{Key: "a", Value: 1}, {Key: "b", Value: 1}, {Key: "c", Value: 1}},
	}
	indexView.CreateOne(ctx, idx)

	numbersCount, _ := numbersCollection.CountDocuments(ctx, bson.M{})
	fmt.Printf("Seeded numbers: %d\n", numbersCount)
	return err
}

// SeedVehicles seeds vehicles collection
func (f *Seed) SeedVehicles(client *mongo.Client) error {
	var err error
	var ctx = context.Background()
	vehiclesCollection := client.Database(f.database).Collection("vehicles")
	dealersCollection := client.Database(f.database).Collection("dealers")
	employeesCollection := client.Database(f.database).Collection("employees")
	if f.isDrop {
		vehiclesCollection.Drop(ctx)
		dealersCollection.Drop(ctx)
		employeesCollection.Drop(ctx)
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

	var emp bson.M
	opts := options.Replace()
	opts.SetUpsert(true)
	var empID = int(1001)
	emp = getEmployee(empID, 0)
	empID++
	employeesCollection.ReplaceOne(ctx, bson.M{"_id": emp["_id"]}, emp, opts)
	for i := 0; i < 2; i++ {
		emp = getEmployee(empID, 1001)
		parent := empID
		employeesCollection.ReplaceOne(ctx, bson.M{"_id": emp["_id"]}, emp, opts)
		empID++
		for j := 0; j < 3; j++ {
			emp = getEmployee(empID, parent)
			pID := empID
			employeesCollection.ReplaceOne(ctx, bson.M{"_id": emp["_id"]}, emp, opts)
			empID++
			for k := 0; k < 5; k++ {
				emp = getEmployee(empID, pID)
				employeesCollection.ReplaceOne(ctx, bson.M{"_id": emp["_id"]}, emp, opts)
				empID++
			}
		}
	}

	// create index example
	indexView := vehiclesCollection.Indexes()
	indexView.CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{"location", 1}, {"color", 1}}})
	indexView.CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "color", Value: 1}}})
	indexView.CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "color", Value: 1}, {Key: "brand", Value: 1}}})
	indexView.CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "dealer", Value: 1}}})
	indexView.CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "dealer", Value: 1}, {Key: "coordinate", Value: "2dsphere"}}})
	indexView.CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "filters.k", Value: 1}, {Key: "filters.v", Value: 1}}})
	dealersCount, _ := dealersCollection.CountDocuments(ctx, bson.M{})
	vehiclesCount := f.seedCollection(vehiclesCollection, 1)
	fopts := options.Find()
	filter := bson.D{{Key: "color", Value: "Red"}}
	fopts.SetSort(bson.D{{Key: "brand", Value: -1}})
	fopts.SetProjection(bson.D{{Key: "_id", Value: 0}, {Key: "color", Value: 1}, {Key: "brand", Value: 11}})
	vehiclesCollection.Find(ctx, filter, fopts)
	fmt.Printf("Seeded vehicles: %d, dealers: %d\n", vehiclesCount, dealersCount)
	return err
}

var dealers = []string{"Atlanta Auto", "Buckhead Auto", "Johns Creek Auto"}
var brands = []string{"Acura", "Alfa Romeo", "Audi", "Bentley", "BMW", "Buick", "Cadillac", "Chevrolet", "Chrysler", "Dodge",
	"Fiat", "Ford", "GMC", "Genesis", "Honda", "Hyundai", "Infiniti", "Jaguar", "Jeep", "Kia",
	"Land Rover", "Lexus", "Lincoln", "Maserati", "Mazda", "Mercedes-Benz", "Nissan", "Porsche", "Toyota", "Volkswagen"}
var styles = []string{"Sedan", "Coupe", "Convertible", "Minivan", "SUV", "Truck"}
var colors = []string{"Beige", "Black", "Blue", "Brown", "Gold",
	"Gray", "Green", "Orange", "Pink", "Purple",
	"Red", "Silver", "White", "Yellow"}
var locations = []string{"US-NY", "US-GA", "US-IL", "US-TX", "US-CA", "US-WA"}

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
		"location": locations[rand.Intn(len(locations))],
		"dealer":   fmt.Sprintf("DEALER-%d", 1+rand.Intn(len(dealers))),
		"brand":    brand,
		"color":    color,
		"style":    style,
		"year":     year,
		"used":     used,
		"filters": []bson.M{
			bson.M{"k": "brand", "v": brand},
			bson.M{"k": "color", "v": color},
			bson.M{"k": "style", "v": style},
			bson.M{"k": "year", "v": year},
			bson.M{"k": "used", "v": used}},
	}
}

func (f *Seed) seedCollection(c *mongo.Collection, fnum int) int {
	fmt.Println("number of connections:", f.conns)
	var ctx = context.Background()
	var bsize = getBatchSize(f.total, f.conns)
	var remaining = f.total

	var wg = gox.NewWaitGroup(f.conns)
	for threadNum := 0; threadNum < f.total; threadNum += bsize {
		wg.Add(1)
		num := bsize
		if remaining < bsize {
			num = remaining
		}
		remaining -= num
		go func(num int) {
			defer wg.Done()
			var contentArray []interface{}
			for n := 0; n < num; n++ {
				if fnum == 1 {
					doc := getVehicle()
					doc["_batch"] = n % 4
					contentArray = append(contentArray, doc)
				} else if fnum == 2 {
					contentArray = append(contentArray, util.GetDemoDoc())
				}
			}
			opts := options.InsertMany()
			opts.SetOrdered(false) // ignore duplication errors
			c.InsertMany(ctx, contentArray, opts)
			if f.showProgress {
				fmt.Fprintf(os.Stderr, "\r%3.1f%% ", float64(100*(f.total-remaining))/float64(f.total))
			}
		}(num)
	}
	wg.Wait()

	if f.showProgress {
		fmt.Fprintf(os.Stderr, "\r        \r")
	}
	cnt, _ := c.CountDocuments(ctx, bson.M{})
	return int(cnt)
}

// SeedFromTemplate seeds data from a template in a file
func (f *Seed) seedFromTemplate(client *mongo.Client) error {
	var err error
	var ctx = context.Background()
	var bsize = getBatchSize(f.total, f.conns)
	var remaining = f.total
	var sdoc bson.M
	var buf []byte
	if sdoc, err = util.GetDocByTemplate(f.file, true); err != nil {
		return err
	}
	if buf, err = json.Marshal(sdoc); err != nil {
		return err
	}
	doc := make(map[string]interface{})
	json.Unmarshal(buf, &doc)
	collName := f.collection
	if collName == "" {
		collName = mdb.ExamplesCollection
	}
	log.Println("Seed data to collection", collName, "using", f.conns, "connections")
	c := client.Database(f.database).Collection(collName)
	var uniq *mdb.Index
	if f.isDrop {
		c.Drop(ctx)
	} else {
		iview, _ := c.Indexes().List(ctx)
		for iview.Next(ctx) {
			var doc mdb.Index
			iview.Decode(&doc)
			if doc.Unique == true {
				uniq = &doc
				break
			}
		}
	}
	if uniq != nil {
		fmt.Println("* unique index detected:", uniq.Name)
		fmt.Print("* keyhole may not be able to seed all data, continue? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)
		if text != "y" && text != "Y" {
			os.Exit(0)
		}
	}

	var mutex sync.RWMutex
	var wg = gox.NewWaitGroup(f.conns)
	var threads int
	var zeroInserted int
	for remaining > 0 && zeroInserted < 3 {
		wg.Add(1)
		num := bsize
		if remaining < bsize {
			num = remaining
		}
		remaining -= num
		threads++
		go func(num int) {
			defer wg.Done()
			inserted, err := populateData(c, num, doc)
			remaining += (num - inserted)
			if err != nil {
				if mdb.IsUnauthorizedError(err) == true {
					return
				}
				// log.Println("bulkWrite failed", err)
				time.Sleep(time.Second)
				mutex.Lock()
				if inserted == 0 {
					zeroInserted++
				}
				mutex.Unlock()
			}
			if f.showProgress {
				fmt.Fprintf(os.Stderr, "\r%3.1f%% ", float64(100*(f.total-remaining))/float64(f.total))
			}
		}(num)
	}
	wg.Wait()
	if remaining > 0 {
		inserted, err := populateData(c, remaining, doc) // catchup
		remaining -= inserted
		if err != nil && mdb.IsUnauthorizedError(err) == true {
			log.Fatal(err)
		}
	}

	if f.showProgress {
		fmt.Fprintf(os.Stderr, "\r        \r")
	}
	fmt.Printf("\rSeeded %s: %d, inserted: %d\n", collName, f.total, (f.total - remaining))
	return err
}

func getEmployee(id int, supervisor int) bson.M {
	dealerID := "DEALER-1"
	email := util.GetEmailAddress()
	s := strings.Split(strings.Split(email, "@")[0], ".")
	doc := bson.M{"_id": int32(id), "dealer": dealerID, "email": email, "name": s[0] + " " + s[2]}
	if supervisor != 0 {
		doc["manager"] = int32(supervisor)
	}
	return doc
}

func getBatchSize(total int, conns int) int {
	size := total / conns
	if total <= 1000 { // for the fun of seeing percentage
		return 100
	} else if size >= 1000 {
		return 1000
	}
	return size
}

func populateData(c *mongo.Collection, num int, doc map[string]interface{}) (int, error) {
	if num == 0 {
		return 0, nil
	}
	var contentArray []interface{}
	for n := 0; n < num; n++ {
		mdoc := make(map[string]interface{})
		util.RandomizeDocument(&mdoc, doc, false)
		contentArray = append(contentArray, mdoc)
	}
	opts := options.InsertMany()
	opts.SetOrdered(false) // ignore _id duplication errors
	res, err := c.InsertMany(context.TODO(), contentArray, opts)
	if err != nil {
		return 0, err
	}
	return len(res.InsertedIDs), err
}
