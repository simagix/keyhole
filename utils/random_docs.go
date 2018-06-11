// Copyright 2018 Kuei-chun Chen. All rights reserved.

package utils

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"
)

// RandomizeDocument traverses a doc and replace values with random values according to their data type.
func RandomizeDocument(doc *map[string]interface{}, f interface{}, meta bool) {
	elems := f.(map[string]interface{})
	for key, value := range elems {
		switch o := value.(type) {
		case map[string]interface{}:
			subdoc := make(map[string]interface{})
			RandomizeDocument(&subdoc, value, meta)
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
			if value.(int) == 1 { // 1 may have special meaning of true
				(*doc)[key] = 1
			} else {
				(*doc)[key] = rand.Intn(10000)
			}
		case float32, float64:
			if value.(float64) == 1 { // 1 may have special meaning of true
				(*doc)[key] = 1
			} else {
				(*doc)[key] = rand.Intn(10000)
			}
		case string:
			if meta == false {
				if value.(string) == "$date" || isDateString(value.(string)) {
					(*doc)[key] = getDate()
				} else if value.(string) == "$oId" || (len(value.(string)) == 24 && isHexString(value.(string))) {
					(*doc)[key] = bson.NewObjectIdWithTime(time.Now())
				} else if isHexString(value.(string)) {
					(*doc)[key] = getHexString(len(value.(string)))
				} else {
					(*doc)[key] = getMagicString(value.(string), meta)
				}
			} else {
				(*doc)[key] = getMagicString(value.(string), meta)
			}
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
			RandomizeDocument(&subdoc1, value, meta)
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
		} else if isHexString(str) && len(str) == 24 {
			return "$oId"
		}
		return str
	}

	if str == "$mail" || isEmailAddress(str) {
		return getEmailAddress()
	} else if str == "$ip" || isIP(str) {
		return getIP()
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
	hexstr := "1234567890ABCDEF"
	hex := ""
	for len(hex) < n {
		hex += string(hexstr[rand.Intn(len(hexstr)-1)])
	}
	return hex
}

func isDateString(str string) bool {
	var matched = regexp.MustCompile(`^\d{4}\-(0?[1-9]|1[012])\-(0?[1-9]|[12][0-9]|3[01])T.*$`)
	return matched.MatchString(str)
}

func getDate() time.Time {
	min := time.Date(1970, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	max := time.Date(2070, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	delta := max - min
	sec := rand.Int63n(delta) + min
	return time.Unix(sec, 0)
}

// PrintQuote print a random quote
func PrintQuote() {
	rand.Seed(time.Now().Unix())
	fmt.Println(quotes[rand.Intn(len(quotes))])
}

// PrintQuotes print all quotes
func PrintQuotes() {
	for _, q := range quotes {
		fmt.Println(q)
	}
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
