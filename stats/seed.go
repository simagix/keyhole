package stats

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const db = "_KEYHOLE_"
const x = 1024

// DBName -
var DBName = fmt.Sprintf("_KEYHOLE_%X", x+x*rand.Intn(x))

// Model -
type Model struct {
	ID          string `json:"_id" bson:"_id"`
	Name        string
	Description string
	Year        int
}

// Task -
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
	session.SetMode(mgo.Monotonic, true)
	session.DB(db).DropDatabase()
	cm := session.DB(db).C("models")
	cr := session.DB(db).C("robots")

	for i := 1000; i < 1050; i++ {
		model := "model-" + fmt.Sprintf("%x", (rand.Intn(5000)+5000)*i)
		name := fmt.Sprintf("Robo %d-%x", i, rand.Intn(1000000))
		descr := fmt.Sprintf("%s %s", model, name)
		year := time.Now().Year() - rand.Intn(5)
		err := cm.Insert(&Model{model, name, descr, year})
		if err != nil {
			log.Fatal(err)
		}

		for r := 0; r < 20+rand.Intn(30); r++ {
			id := "robot-" + fmt.Sprintf("%x", (rand.Intn(5000)+5000)*r)
			notes := fmt.Sprintf("%s %s", id, model)
			pct := rand.Float32()
			tasks := []Task{{"Business", 10 + rand.Intn(60)}, {"Home", 10 + rand.Intn(60)}}
			err := cr.Insert(&Robot{id, model, notes, pct, tasks})
			if err != nil {
				continue
			}
		}

		if verbose == true {
			modelRes := Model{}
			robotRes := []Robot{}
			_ = cm.Find(bson.M{"_id": model}).One(&modelRes)
			fmt.Println(modelRes)
			_ = cr.Find(bson.M{"modelId": model}).All(&robotRes)
			fmt.Println(robotRes)
		}
	}
	mn, _ := cm.Count()
	rn, _ := cr.Count()
	fmt.Printf("Seeded models: %d, robots: %d\n", mn, rn)

	isNew := []bool{true, false}
	styles := []string{"Sedan", "Coupe", "Convertible", "Minivan", "SUV", "Truck"}
	colors := []string{"Beige", "Black", "Blue", "Brown", "Gold", "Gray", "Green", "Orange", "Pink", "Purple", "Red", "Silver", "White", "Yellow"}

	c := session.DB(db).C("cars")
	for i := 0; i < 250; i++ {
		bulk := c.Bulk()
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
	abc := session.DB(db).C("numbers")
	bulk := abc.Bulk()
	for n := 0; n < 100000; n++ {
		contentArray = append(contentArray, bson.M{"a": rand.Intn(100), "b": rand.Intn(50), "c": rand.Intn(1000)})
	}
	bulk.Insert(contentArray...)
	_, err := bulk.Run()
	if err != nil {
		log.Println(err)
		return
	}
}
