// Copyright 2018 Kuei-chun Chen. All rights reserved.

package stats

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
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
	traverseDocument(&doc, f, meta)
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

// SeedFromTemplate seeds data from a template in a file
func SeedFromTemplate(session *mgo.Session, filename string, isDrop bool, verbose bool) {
	sdoc := GetDocByTemplate(filename, true)
	bytes, _ := json.MarshalIndent(sdoc, "", "   ")
	if verbose {
		fmt.Println(string(bytes))
	}
	doc := make(map[string]interface{})
	json.Unmarshal(bytes, &doc)
	session.SetMode(mgo.Primary, true)
	var contentArray []interface{}
	examplesCollection := session.DB(keyholeDB).C("examples")
	if isDrop {
		examplesCollection.DropCollection()
	}
	bulk := examplesCollection.Bulk()
	total := 1000
	for n := 0; n < total; n++ {
		ndoc := make(map[string]interface{})
		fmt.Fprintf(os.Stderr, "\r%3d%% ", 100*n/total)
		traverseDocument(&ndoc, doc, false)
		delete(ndoc, "_id")
		contentArray = append(contentArray, ndoc)
	}
	fmt.Fprintf(os.Stderr, "\r100%% \n")
	bulk.Insert(contentArray...)
	_, err := bulk.Run()
	if err != nil {
		log.Println(err)
		return
	}
	examplesCount, _ := examplesCollection.Count()
	fmt.Printf("\rSeeded examples: %d, total count: %d\n", total, examplesCount)
}

// traverse a doc and replace values with random values according to their data type.
func traverseDocument(doc *map[string]interface{}, f interface{}, meta bool) {
	elems := f.(map[string]interface{})
	for key, value := range elems {
		switch o := value.(type) {
		case map[string]interface{}:
			subdoc := make(map[string]interface{})
			traverseDocument(&subdoc, value, meta)
			(*doc)[key] = subdoc
		case []interface{}:
			subdoc := make([]interface{}, len(o))
			getArrayOfRandomDocs(o, &subdoc, meta)
			(*doc)[key] = subdoc
		case bool:
			b := true
			if rand.Int()%2 == 0 {
				b = false
			}
			(*doc)[key] = b
		case int, int8, int16, int32, int64:
			(*doc)[key] = rand.Intn(10000)
		case float32, float64:
			(*doc)[key] = rand.Intn(10000)
		case string:
			(*doc)[key] = getMagicString(value.(string), meta)
		default:
			(*doc)[key] = value
		}
	}
}

func getArrayOfRandomDocs(obj []interface{}, subdoc *[]interface{}, meta bool) {
	for key, value := range obj {
		switch o := value.(type) {
		case bool:
			b := true
			if rand.Int()%2 == 0 {
				b = false
			}
			(*subdoc)[key] = b
		case int, int8, int16, int32, int64:
			(*subdoc)[key] = rand.Intn(10000)
		case float32, float64:
			(*subdoc)[key] = rand.Intn(10000)
		case string:
			(*subdoc)[key] = getMagicString(value.(string), meta)
		case []interface{}:
			subdocument := make([]interface{}, len(o))
			getArrayOfRandomDocs(o, &subdocument, meta)
			(*subdoc)[key] = subdocument
		case map[string]interface{}:
			subdoc1 := make(map[string]interface{})
			traverseDocument(&subdoc1, value, meta)
			(*subdoc)[key] = subdoc1
		case interface{}:
			fmt.Println("=>", o)
			(*subdoc)[key] = value
		default:
		}
	}
}

// Returns randomized string.  if meta is true, it intends to avoid future regex
// actions by replacing the values with $mail, $ip, and $date.
func getMagicString(str string, meta bool) string {
	if meta {
		if isEmailAddress(str) {
			return "$mail"
		} else if isIP(str) {
			return "$ip"
		} else if isDateString(str) {
			return "$date"
		}
		return str
	}

	if str == "$mail" || isEmailAddress(str) {
		return getEmailAddress()
	} else if str == "$ip" || isIP(str) {
		return getIP()
	} else if str == "$date" || isDateString(str) {
		return getDateString(len(str))
	} else if isHexString(str) {
		return getHexString(len(str))
	}

	if len(str) < 10 {
		return fnames[rand.Intn(len(fnames))]
	}
	quote := ""
	for len(quote) < len(str) {
		quote += quotes[rand.Intn(len(quotes))] + " "
	}
	quote = quote[:len(str)]
	quote = strings.Trim(quote, " ")
	idx := strings.LastIndex(quote, " ")
	if idx < 0 {
		return quote
	}

	return quote[:idx]
}

func isEmailAddress(str string) bool {
	var matched = regexp.MustCompile(`^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+$`)
	return matched.MatchString(str)
}

func getEmailAddress() string {
	return fnames[rand.Intn(len(fnames)-1)] + "." +
		string(fnames[rand.Intn(len(fnames)-1)][0]) + "." +
		lnames[rand.Intn(len(lnames)-1)] + "@" +
		domains[rand.Intn(len(domains)-1)]
}

func isIP(str string) bool {
	var matched = regexp.MustCompile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)
	return matched.MatchString(str)
}

func getIP() string {
	return strconv.Itoa(rand.Intn(255)) + "." + strconv.Itoa(rand.Intn(255)) + "." +
		strconv.Itoa(rand.Intn(255)) + "." + strconv.Itoa(rand.Intn(255))
}

func isHexString(str string) bool {
	var matched = regexp.MustCompile(`^[\da-fA-F]+$`)
	return matched.MatchString(str)
}

func getHexString(n int) string {
	hexstr := "1234567890abcdef"
	hex := ""
	for len(hex) < n {
		hex += string(hexstr[rand.Intn(len(hexstr)-1)])
	}
	return hex
}

func isDateString(str string) bool {
	var matched = regexp.MustCompile(`^\d{4}\-(0?[1-9]|1[012])\-(0?[1-9]|[12][0-9]|3[01])T.*Z$`)
	return matched.MatchString(str)
}

func getDateString(n int) string {
	min := time.Date(1970, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	max := time.Date(2070, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	delta := max - min
	sec := rand.Int63n(delta) + min
	return time.Unix(sec, 0).Format(time.RFC3339)
}

var quotes = []string{
	"Frankly, my dear, I don't give a damn.",
	"I'm going to make him an offer he can't refuse.",
	"Toto, I've a feeling we're not in Kansas anymore.",
	"Here's looking at you, kid.",
	"Go ahead, make my day.",
	"All right, Mr. DeMille, I'm ready for my close-up.",
	"May the Force be with you.",
	"Fasten your seatbelts. It's going to be a bumpy night.",
	"You talkin' to me?",
	"What we've got here is failure to communicate.",
	"I love the smell of napalm in the morning.",
	"Love means never having to say you're sorry.",
	"The stuff that dreams are made of.",
	"E.T. phone home.",
	"They call me Mister Tibbs!",
	"You're gonna need a bigger boat.",
	"Of all the gin joints in all the towns in all the world, she walks into mine.",
	"Bond. James Bond.",
	"There's no place like home.",
	"Show me the money!",
}

var domains = []string{"gmail.com", "me.com", "yahoo.com", "outlook.com", "google.com",
	"simagix.com", "aol.com", "mongodb.com", "example.com", "cisco.com",
	"microsoft.com", "facebook.com", "apple.com", "amazon.com", "oracle.com"}
var fnames = []string{"Liam", "Emma", "Noah", "Olivia", "Willaim",
	"Ava", "James", "Isabella", "Logan", "Sophia",
	"John", "Robert", "Michael", "David", "Richard",
	"Mary", "Patricia", "Jennifer", "Linda", "Elizabeth"}
var lnames = []string{"Smith", "Johnson", "Williams", "Brown", "Jones",
	"Miller", "Davis", "Garcia", "Rodriguez", "Chen",
	"Adams", "Arthur", "Bush", "Carter", "Clinton",
	"Eisenhower", "Ford", "Grant", "Harrison", "Hoover"}
