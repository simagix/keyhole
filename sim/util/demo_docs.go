// Copyright 2018 Kuei-chun Chen. All rights reserved.

package util

import (
	"math/rand"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

var (
	locations = []string{"US-NY", "US-GA", "US-IL", "US-TX", "US-CA", "US-WA"}
)

// FavoritesSchema -
type FavoritesSchema struct {
	ID            string `json:"_id" bson:"_id"`
	FavoriteBook  string `json:"favoriteBook" bson:"favoriteBook"`
	FavoriteCity  string `json:"favoriteCity" bson:"favoriteCity"`
	FavoriteMovie string `json:"favoriteMovie" bson:"favoriteMovie"`
	FavoriteMusic string `json:"favoriteMusic" bson:"favoriteMusic"`
	FavoriteSport string `json:"favoriteSport" bson:"favoriteSport"`
}

// FavoritesDoc -
type FavoritesDoc struct {
	Sports []string
	Music  []string
	Cities []string
	Books  []string
	Movies []string
}

// GetDemoDoc returns a demo document
func GetDemoDoc() bson.M {
	var n = len(Favorites.Sports)
	favoriteSports := []string{Favorites.Sports[rand.Intn(n)], Favorites.Sports[rand.Intn(n)], Favorites.Sports[rand.Intn(n)]}
	favoriteSports = unique(append(favoriteSports, Favorites.Sports[0:3]...), 3)
	n = len(Favorites.Music)
	favoriteMusic := []string{Favorites.Music[rand.Intn(n)], Favorites.Music[rand.Intn(n)], Favorites.Music[rand.Intn(n)]}
	favoriteMusic = unique(append(favoriteMusic, Favorites.Music[0:3]...), 3)
	n = len(Favorites.Cities)
	favoriteCities := []string{Favorites.Cities[rand.Intn(n)], Favorites.Cities[rand.Intn(n)], Favorites.Cities[rand.Intn(n)]}
	favoriteCities = unique(append(favoriteCities, Favorites.Cities[0:3]...), 3)
	n = len(Favorites.Books)
	x := rand.Intn(n)
	favoriteBooks := []string{Favorites.Books[x], Favorites.Books[rand.Intn(n)], Favorites.Books[rand.Intn(n)]}
	favoriteBooks = unique(append(favoriteBooks, Favorites.Books[0:3]...), 3)
	n = len(Favorites.Movies)
	favoriteMovies := []string{Favorites.Movies[rand.Intn(n)], Favorites.Movies[rand.Intn(n)], Favorites.Movies[rand.Intn(n)]}
	favoriteMovies = unique(append(favoriteMovies, Favorites.Movies[0:3]...), 3)
	favoritesList := []bson.M{
		{"sport": favoriteSports[0], "music": favoriteMusic[0], "city": favoriteCities[0], "book": favoriteBooks[0], "movie": favoriteMovies[0]},
		{"sport": favoriteSports[1], "music": favoriteMusic[1], "city": favoriteCities[1], "book": favoriteBooks[1], "movie": favoriteMovies[1]},
		{"sport": favoriteSports[2], "music": favoriteMusic[2], "city": favoriteCities[2], "book": favoriteBooks[2], "movie": favoriteMovies[2]},
	}
	favoritesKVList := []bson.M{}
	for i := 0; i < 3; i++ {
		favoritesKVList = append(favoritesKVList,
			bson.M{
				"level": int32(i + 1),
				"categories": []bson.M{
					{"key": "sport", "value": favoriteSports[i]},
					{"key": "music", "value": favoriteMusic[i]},
					{"key": "city", "value": favoriteCities[i]},
					{"key": "book", "value": favoriteBooks[i]},
					{"key": "movie", "value": favoriteMovies[i]},
				}})
	}
	favoritesKVSet := []bson.M{}
	favoritesKVSet = append(favoritesKVSet, bson.M{"key": "sport", "value": favoriteSports})
	favoritesKVSet = append(favoritesKVSet, bson.M{"key": "music", "value": favoriteMusic})
	favoritesKVSet = append(favoritesKVSet, bson.M{"key": "city", "value": favoriteCities})
	favoritesKVSet = append(favoritesKVSet, bson.M{"key": "book", "value": favoriteBooks})
	favoritesKVSet = append(favoritesKVSet, bson.M{"key": "movie", "value": favoriteMovies})

	email := GetEmailAddress()
	s := strings.Split(strings.Split(email, "@")[0], ".")
	doc := bson.M{
		"location":  locations[rand.Intn(len(locations))],
		"_search":   strconv.FormatInt(rand.Int63(), 16),
		"email":     email,
		"firstName": s[0],
		"lastName":  s[2],
		"favoritesAll": bson.M{
			"sports": favoriteSports,
			"musics": favoriteMusic,
			"cities": favoriteCities,
			"books":  favoriteBooks,
			"movies": favoriteMovies,
		},
		"favoritesList":   favoritesList,
		"favoritesKVList": favoritesKVList,
		"favoritesKVSet":  favoritesKVSet,
		"favoriteBookId":  1100 + x,
		"favoriteBook":    favoriteBooks[0],
		"favoriteBooks":   favoriteBooks,
		"favoriteCity":    favoriteCities[0],
		"favoriteCities":  favoriteCities,
		"favoriteMovie":   favoriteMovies[0],
		"favoriteMovies":  favoriteMovies,
		"favoriteMusic":   favoriteMusic[0],
		"favoriteMusics":  favoriteMusic,
		"favoriteSport":   favoriteSports[0],
		"favoriteSports":  favoriteSports,
		"number":          rand.Intn(1000),
		"numbers":         []int{rand.Intn(1000), rand.Intn(1000), rand.Intn(1000), rand.Intn(1000), rand.Intn(1000)},
		"ts":              time.Now(),
	}
	return doc
}

func unique(s []string, n int) []string {
	arr := make(map[string]bool, len(s))
	us := make([]string, len(arr))
	for _, elem := range s {
		if len(elem) != 0 {
			if !arr[elem] {
				us = append(us, elem)
				arr[elem] = true
			}
		}
	}
	return us[:n]
}

// Favorites constance
var Favorites = FavoritesDoc{
	Sports: []string{"Baseball", "Boxing", "Dodgeball", "Figure skating", "Football", "Horse racing", "Mountaineering", "Skateboard", "Ski", "Soccer"},
	Music:  []string{"Blues", "Classical", "Country", "Easy Listening", "Electronic", "Hip Pop", "Jazz", "Opera", "Soul", "Rock"},
	Cities: []string{"Atlanta", "Bangkok", "Beijing", "London", "Paris", "Singapore", "New York", "Istanbul", "Dubai", "Taipei"},
	Books:  []string{"Journey to the West", "The Lord of the Rings", "In Search of Lost Time", "Don Quixote", "Ulysses", "The Great Gatsby", "Moby Dick", "Hamlet", "War and Peace", "The Odyssey"},
	Movies: []string{"The Godfather", "The Shawshank Redemption", "Schindler's List", "Raging Bull", "Casablanca", "Citizen Kane", "Gone with the Wind", "The Wizard of Oz", "One Flew Over the Cuckoo's Nest", "Lawrence of Arabia"},
}
