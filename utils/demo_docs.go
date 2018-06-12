// Copyright 2018 Kuei-chun Chen. All rights reserved.

package utils

import (
	"bytes"
	"math/rand"
	"strconv"
	"time"

	"gopkg.in/mgo.v2/bson"
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

// Favorites -
type Favorites struct {
	Sports []string
	Music  []string
	Cities []string
	Books  []string
	Movies []string
}

// GetDemoDoc returns a demo document
func GetDemoDoc() bson.M {
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
		"_search": strconv.FormatInt(rand.Int63(), 16),
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

var favorites = Favorites{
	Sports: []string{"Baseball", "Boxing", "Dodgeball", "Figure skating", "Football", "Horse racing", "Mountaineering", "Skateboard", "Ski", "Soccer"},
	Music:  []string{"Blues", "Classical", "Country", "Easy Listening", "Electronic", "Hip Pop", "Jazz", "Opera", "Soul", "Rock"},
	Cities: []string{"Atlanta", "Bangkok", "Beijing", "London", "Paris", "Singapore", "New York", "Istanbul", "Dubai", "Taipei"},
	Books:  []string{"Journey to the West", "The Lord of the Rings", "In Search of Lost Time", "Don Quixote", "Ulysses", "The Great Gatsby", "Moby Dick", "Hamlet", "War and Peace", "The Odyssey"},
	Movies: []string{"The Godfather", "The Shawshank Redemption", "Schindler's List", "Raging Bull", "Casablanca", "Citizen Kane", "Gone with the Wind", "The Wizard of Oz", "One Flew Over the Cuckoo's Nest", "Lawrence of Arabia"},
}
