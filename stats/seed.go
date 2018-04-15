package stats

import (
	"fmt"
	"log"
	mrand "math/rand"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const db = "_KEYHOLE_"

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
func Seed(uri string, verbose bool) {
	session, err := mgo.Dial(uri)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)
	session.DB(db).DropDatabase()
	cm := session.DB(db).C("models")
	cr := session.DB(db).C("robots")

	for i := 1000; i < 1050; i++ {
		model := "model-" + fmt.Sprintf("%x", (mrand.Intn(5000)+5000)*i)
		name := fmt.Sprintf("Robo %d-%x", i, mrand.Intn(1000000))
		descr := fmt.Sprintf("%s %s", model, name)
		year := time.Now().Year() - mrand.Intn(5)
		err = cm.Insert(&Model{model, name, descr, year})
		if err != nil {
			log.Fatal(err)
		}

		for r := 0; r < 20+mrand.Intn(30); r++ {
			id := "robot-" + fmt.Sprintf("%x", (mrand.Intn(5000)+5000)*r)
			notes := fmt.Sprintf("%s %s", id, model)
			pct := mrand.Float32()
			tasks := []Task{{"Business", 10 + mrand.Intn(60)}, {"Home", 10 + mrand.Intn(60)}}
			err = cr.Insert(&Robot{id, model, notes, pct, tasks})
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
}
