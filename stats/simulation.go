package stats

import (
	"bytes"
	"encoding/json"
	"log"
	"math"
	"math/rand"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// CollectionName -
var CollectionName = "keyhole"

// Schema -
type Schema struct {
	ID            string `json:"_id",bson:"_id"`
	FavoriteCity  string `json:"favoriteCity",bson:"favoriteCity"`
	FavoriteBook  string `json:"favoriteBook",bson:"favoriteBook"`
	FavoriteMovie string `json:"favoriteMovie",bson:"favoriteMovie"`
}

// Favorites -
type Favorites struct {
	Sports []string
	Music  []string
	Cities []string
	Books  []string
	Movies []string
}

var favorites = Favorites{
	Sports: []string{"Baseball", "Boxing", "Dodgeball", "Figure skating", "Football", "Horse racing", "Mountaineering", "Skateboard", "Ski", "Soccer"},
	Music:  []string{"Blues", "Classical", "Country", "Easy Listening", "Electronic", "Hip Pop", "Jazz", "Opera", "Soul", "Rock"},
	Cities: []string{"Atlanta", "Bangkok", "Beijing", "London", "Paris", "Singapore", "New York", "Istanbul", "Dubai", "Taipei"},
	Books:  []string{"Journey to the West", "The Lord of the Rings", "In Search of Lost Time", "Don Quixote", "Ulysses", "The Great Gatsby", "Moby Dick", "Hamlet", "War and Peace", "The Odyssey"},
	Movies: []string{"The Godfather", "The Shawshank Redemption", "Schindler's List", "Raging Bull", "Casablanca", "Citizen Kane", "Gone with the Wind", "The Wizard of Oz", "One Flew Over the Cuckoo's Nest", "Lawrence of Arabia"},
}

// MongoConn -
type MongoConn struct {
	uri    string
	ssl    bool
	sslCA  string
	dbName string
	tps    int
}

// New - Constructor
func New(uri string, ssl bool, sslCA string, dbName string, tps int) MongoConn {
	m := MongoConn{uri, ssl, sslCA, dbName, tps}
	return m
}

func unique(s []string) []string {
	unique := make(map[string]bool, len(s))
	us := make([]string, len(unique))
	for _, elem := range s {
		if len(elem) != 0 {
			if !unique[elem] {
				us = append(us, elem)
				unique[elem] = true
			}
		}
	}
	return us[:3]
}

// GetRandomDoc -
func GetRandomDoc() bson.M {
	var filler1, filler2 bytes.Buffer
	for i := 0; i < 80/len("simagix."); i++ {
		filler1.WriteString("simagix.")
		filler2.WriteString("mongodb.")
	}
	var n = len(favorites.Sports)
	favoriteSports := []string{favorites.Sports[rand.Intn(n)], favorites.Sports[rand.Intn(n)], favorites.Sports[rand.Intn(n)]}
	favoriteSports = unique(append(favoriteSports, favorites.Sports[0:3]...))
	n = len(favorites.Music)
	favoriteMusic := []string{favorites.Music[rand.Intn(n)], favorites.Music[rand.Intn(n)], favorites.Music[rand.Intn(n)]}
	favoriteMusic = unique(append(favoriteMusic, favorites.Music[0:3]...))
	n = len(favorites.Cities)
	favoriteCities := []string{favorites.Cities[rand.Intn(n)], favorites.Cities[rand.Intn(n)], favorites.Cities[rand.Intn(n)]}
	favoriteCities = unique(append(favoriteCities, favorites.Cities[0:3]...))
	n = len(favorites.Books)
	favoriteBooks := []string{favorites.Books[rand.Intn(n)], favorites.Books[rand.Intn(n)], favorites.Books[rand.Intn(n)]}
	favoriteBooks = unique(append(favoriteBooks, favorites.Books[0:3]...))
	n = len(favorites.Movies)
	favoriteMovies := []string{favorites.Movies[rand.Intn(n)], favorites.Movies[rand.Intn(n)], favorites.Movies[rand.Intn(n)]}
	favoriteMovies = unique(append(favoriteMovies, favorites.Movies[0:3]...))
	favoritesList := []bson.M{
		bson.M{"sport": favoriteSports[0], "music": favoriteMusic[0], "city": favoriteCities[0], "book": favoriteBooks[0], "movie": favoriteMovies[0]},
		bson.M{"sport": favoriteSports[1], "music": favoriteMusic[1], "city": favoriteCities[1], "book": favoriteBooks[1], "movie": favoriteMovies[1]},
		bson.M{"sport": favoriteSports[2], "music": favoriteMusic[2], "city": favoriteCities[2], "book": favoriteBooks[2], "movie": favoriteMovies[2]},
	}

	doc := bson.M{
		"favorites": bson.M{
			"sports": favoriteSports, "sport": favoriteSports[0],
			"musics": favoriteMusic, "music": favoriteMusic[0],
			"cities": favoriteCities, "city": favoriteCities[0],
			"books": favoriteBooks, "book": favoriteBooks[0],
			"movies": favoriteMovies, "movie": favoriteMovies[0],
		},
		"favoritesList":  favoritesList,
		"favoriteSports": favoriteSports,
		"favoriteMusics": favoriteMusic,
		"favoriteCities": favoriteCities,
		"favoriteBooks":  favoriteBooks,
		"favoriteMovies": favoriteMovies,
		"favoriteSport":  favoriteSports[0],
		"favoriteMusic":  favoriteMusic[0],
		"favoriteCity":   favoriteCities[0],
		"favoriteBook":   favoriteBooks[0],
		"favoriteMovie":  favoriteMovies[0],
		"filler1":        filler1.String(),
		"filler2":        filler2.String(),
		"number":         rand.Intn(1000),
		"numbers":        []int{rand.Intn(1000), rand.Intn(1000), rand.Intn(1000), rand.Intn(1000), rand.Intn(1000)},
		"ts":             time.Now(),
	}
	return doc
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
func (m MongoConn) PopulateData() {
	rand.Seed(time.Now().Unix())
	s := 0
	batchSize := 20
	if m.tps < batchSize {
		batchSize = m.tps
	}
	for s < 60 {
		s++
		session, err := GetSession(m.uri, m.ssl, m.sslCA)
		if err == nil {
			session.SetMode(mgo.Primary, true)
			c := session.DB(m.dbName).C(CollectionName)
			bt := time.Now()
			bulk := c.Bulk()

			for i := 0; i < m.tps; i += batchSize {
				var contentArray []interface{}
				for n := 0; n < batchSize; n++ {
					contentArray = append(contentArray, GetRandomDoc())
				}
				bulk.Insert(contentArray...)
				_, err := bulk.Run()
				if err != nil {
					log.Println(err)
					session.Close()
					break
				}
			}

			t := time.Now()
			elapsed := t.Sub(bt)
			if elapsed.Seconds() > time.Second.Seconds() {
				x := math.Floor(elapsed.Seconds())
				s += int(x)
				elapsed = time.Duration(elapsed.Seconds() - x)
			}
			et := time.Second.Seconds() - elapsed.Seconds()
			time.Sleep(time.Duration(et))
			session.Close()
		} else {
			time.Sleep(time.Second)
		}
	}
}

// Simulate - Simulate CRUD for load tests
func (m MongoConn) Simulate(duration int) {
	var schema = Schema{}
	results := []bson.M{}
	change := bson.M{"$set": bson.M{"ts": time.Now()}}
	burstBegin := time.NewTimer(1 * time.Minute)
	waitms := 5
	isTeardown := false

	go func() {
		<-burstBegin.C
		waitms = 0
	}()
	burstEnd := time.NewTimer(time.Duration(duration-1) * time.Minute)
	go func() {
		<-burstEnd.C
		waitms = 1
		isTeardown = true
	}()

	for {
		session, err := GetSession(m.uri, m.ssl, m.sslCA)
		session.SetMode(mgo.Primary, true)
		if err == nil {
			c := session.DB(m.dbName).C(CollectionName)
			var book string
			var city string
			var movie string
			for i := 0; i < 500; i++ {
				doc := GetRandomDoc()
				bytes, _ := json.Marshal(doc)
				json.Unmarshal(bytes, &schema)
				city = schema.FavoriteCity
				book = schema.FavoriteBook
				movie = schema.FavoriteMovie

				if isTeardown {
					c.RemoveAll(bson.M{"favoriteCity": city, "favoriteBook": book})
				} else {
					_id := bson.NewObjectIdWithTime(time.Now())
					c.Upsert(_id, doc)
					c.Find(bson.M{"favoriteCity": city}).Sort("favoriteCity").Limit(512).All(&results)
					c.Find(bson.M{"favoriteCity": city, "favoriteBook": book}).One(&results)
					c.Update(bson.M{"_id": _id}, change)
					c.Find(bson.M{"favoriteCity": city, "favoriteBook": book, "FavoriteMovie": movie}).One(&results)
					c.Find(bson.M{"favoritesList": bson.M{"$elemMatch": bson.M{"movie": movie}}}).One(&results)
					c.Find(bson.M{"favoritesList": bson.M{"$elemMatch": bson.M{"book": book}}}).Limit(100).All(&results)
					c.Remove(bson.M{"_id": _id})
				}
				time.Sleep(time.Millisecond * time.Duration(waitms))
			}
			c.Find(bson.M{"favoritesList": bson.M{"$elemMatch": bson.M{"book": book}}}).Sort("favoriteBook").Limit(100).All(&results)
			session.Close()
		}
	}
}

// Cleanup - Drop the temp database
func (m MongoConn) Cleanup() {
	log.Println("cleanup", m.uri)
	session, err := GetSession(m.uri, m.ssl, m.sslCA)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	time.Sleep(2 * time.Second)
	log.Println("dropping collection", m.dbName, CollectionName)
	session.DB(m.dbName).C(CollectionName).DropCollection()
	log.Println("dropping database", m.dbName)
	session.DB(m.dbName).DropDatabase()
}

// CreateIndexes -
func (m MongoConn) CreateIndexes() {
	session, _ := GetSession(m.uri, m.ssl, m.sslCA)
	c := session.DB(m.dbName).C(CollectionName)
	c.EnsureIndexKey("favoriteCity")
}
