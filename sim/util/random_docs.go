// Copyright 2018 Kuei-chun Chen. All rights reserved.

package util

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetDocByTemplate returns a bson.M document
func GetDocByTemplate(filename string, meta bool) (bson.M, error) {
	var buf []byte
	var err error

	if buf, err = ioutil.ReadFile(filename); err != nil {
		return nil, err
	}
	return GetRandomizedDoc(buf, meta)
}

// GetRandomizedDoc returns a randomized doc from byte string
func GetRandomizedDoc(buf []byte, meta bool) (bson.M, error) {
	var err error
	var str string
	re := regexp.MustCompile(`ObjectId\(\S+\)`)
	str = re.ReplaceAllString(string(buf), "\"$$oId\"")
	re = regexp.MustCompile(`NumberDecimal\("?([-+]?[0-9]*\.?[0-9]*)"?\)`)
	str = re.ReplaceAllString(str, "{\"$$numberDecimal\": $1}")
	re = regexp.MustCompile(`NumberLong\("?([-+]?[0-9]*\.?[0-9]*)"?\)`)
	str = re.ReplaceAllString(str, "{\"$$numberLong\": $1}")
	re = regexp.MustCompile(`NumberInt\("?([-+]?[0-9]*\.?[0-9]*)"?\)`)
	str = re.ReplaceAllString(str, "{\"$$numberInt\": $1}")
	re = regexp.MustCompile(`ISODate\(\S+\)`)
	str = re.ReplaceAllString(str, "\"$$date\"")
	re = regexp.MustCompile(`Number\("?(\d+)"?\)`)
	str = re.ReplaceAllString(str, "$1")
	re = regexp.MustCompile(`^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+$`)
	str = re.ReplaceAllString(str, "\"$$email\"")
	var f interface{}
	if err = json.Unmarshal([]byte(str), &f); err != nil {
		return nil, err
	}
	doc := make(map[string]interface{})
	RandomizeDocument(&doc, f, meta)
	return doc, err
}

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
			randBool = !randBool
			(*doc)[key] = randBool
		case int, int8, int16, int32, int64:
			if value.(int) == 1 || value.(int) == 0 { // 1 may have special meaning of true
				(*doc)[key] = value
			} else {
				(*doc)[key] = getNumber(value)
			}
		case float32, float64:
			(*doc)[key] = getNumber(value)
		case string:
			if meta == false {
				if value.(string) == metaDate || isDateString(value.(string)) {
					(*doc)[key] = getDate()
					continue
				} else if value.(string) == metaOID || (len(value.(string)) == 24 && isHexString(value.(string))) {
					(*doc)[key] = primitive.NewObjectID()
					continue
				}
			}
			(*doc)[key] = getMagicString(value.(string), meta)
		default:
			(*doc)[key] = value
		}
	}
}

var randBool bool

func getArrayOfRandomDocs(obj []interface{}, doc *[]interface{}, meta bool) {
	for key, value := range obj {
		switch o := value.(type) {
		case bool:
			randBool = !randBool
			(*doc)[key] = randBool
		case int, int8, int16, int32, int64:
			if value.(int) == 1 || value.(int) == 0 { // 1 may have special meaning of true
				(*doc)[key] = value
			} else {
				(*doc)[key] = getNumber(value)
			}
		case float32, float64:
			(*doc)[key] = getNumber(value)
		case string:
			if meta == false {
				if value.(string) == metaDate || isDateString(value.(string)) {
					(*doc)[key] = getDate()
					continue
				} else if value.(string) == metaOID || (len(value.(string)) == 24 && isHexString(value.(string))) {
					(*doc)[key] = primitive.NewObjectID()
					continue
				}
			}
			(*doc)[key] = getMagicString(value.(string), meta)
		case []interface{}:
			subdocument := make([]interface{}, len(o))
			getArrayOfRandomDocs(o, &subdocument, meta)
			(*doc)[key] = subdocument
		case map[string]interface{}:
			subdoc1 := make(map[string]interface{})
			RandomizeDocument(&subdoc1, value, meta)
			(*doc)[key] = subdoc1
		default:
			(*doc)[key] = value
		}
	}
}

const metaEmail = "$email"
const metaIP = "$ip"
const metaSSN = "$ssn"
const metaTEL = "$tel"
const metaDate = "$date"
const metaOID = "$oId"

// Returns randomized string.  if meta is true, it intends to avoid future regex
// actions by replacing the values with $email, $ip, and $date.
func getMagicString(str string, meta bool) string {
	if meta == true {
		if str == metaEmail || isEmailAddress(str) {
			return metaEmail
		} else if str == metaIP || isIP(str) {
			return metaIP
			// } else if str == metaSSN || isSSN(str) {
			// 	return metaSSN
			// } else if str == metaTEL || isPhoneNumber(str) {
			// 	return metaTEL
		} else if str == metaDate || isDateString(str) {
			return metaDate
		} else if str == metaOID || (len(str) == 24 && isHexString(str)) {
			return metaOID
		}
	} else if str == metaIP || isIP(str) {
		return getIP()
	} else if str == metaEmail || isEmailAddress(str) {
		return GetEmailAddress()
		// } else if str == metaSSN || isSSN(str) {
		// 	return getSSN()
		// } else if str == metaTEL || isPhoneNumber(str) {
		// 	return getPhoneNumber()
		// } else if isHexString(str) {
		// 	return getHexString(len(str))
	} else if strings.HasPrefix(str, "$") { // could be a variable
		return str
	}
	hdr := ""
	if n := strings.Index(str, "://"); n > 0 {
		hdr = str[:n+3]
	}
	b := make([]byte, len(str))
	for i, c := range str {
		x := rand.Int()
		if i < len(hdr) {
			b[i] = byte(c)
		} else if c >= 48 && c <= 57 { // digits
			b[i] = byte(x%10 + 48)
		} else if c >= 65 && c <= 90 { // A-Z
			b[i] = byte(x%26 + 65)
		} else if c >= 97 && c <= 122 { // a-z
			b[i] = byte(x%26 + 97)
		} else {
			b[i] = byte(c)
		}
	}
	return string(b)
}

func isEmailAddress(str string) bool {
	var matched = regexp.MustCompile(`^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+$`)
	return matched.MatchString(str)
}

// GetEmailAddress exposes getEmailAddress()
func GetEmailAddress() string {
	return fmt.Sprintf("%s.%s.%s@%s",
		fnames[rand.Intn(len(fnames)-1)], string(fnames[rand.Intn(len(fnames)-1)][0]),
		lnames[rand.Intn(len(lnames)-1)], domains[rand.Intn(len(domains)-1)])
}

func isIP(str string) bool {
	var matched = regexp.MustCompile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)
	return matched.MatchString(str)
}

func getIP() string {
	return strconv.Itoa(rand.Intn(255)) + "." + strconv.Itoa(rand.Intn(255)) + "." +
		strconv.Itoa(rand.Intn(255)) + "." + strconv.Itoa(rand.Intn(255))
}

func isSSN(str string) bool {
	var matched = regexp.MustCompile(`^(\d{3}-?\d{2}-?\d{4}|XXX-XX-XXXX)$`)
	return matched.MatchString(str)
}

func getSSN() string {
	return fmt.Sprintf("%v-%v-5408", (100 + rand.Intn(899)), (10 + rand.Intn(89)))
}

func isPhoneNumber(str string) bool {
	var matched = regexp.MustCompile(`^(\+\d{1,2}\s)?\(?\d{3}\)?[\s.-]\d{3}[\s.-]\d{4}$`)
	return matched.MatchString(str)
}

func getPhoneNumber() string {
	return fmt.Sprintf("(%v) 555-%v", (100 + rand.Intn(899)), (1000 + rand.Intn(8999)))
}

func isHexString(str string) bool {
	var matched = regexp.MustCompile(`^[\da-fA-F]+$`)
	return matched.MatchString(str)
}

func getHexString(n int) string {
	bytes := make([]byte, n/2)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}

func isDateString(str string) bool {
	var matched = regexp.MustCompile(`^\d{4}\-(0?[1-9]|1[012])\-(0?[1-9]|[12][0-9]|3[01])T.*$`)
	return matched.MatchString(str)
}

var now = time.Now()
var min = now.AddDate(-1, 0, 0).Unix()
var max = now.AddDate(0, 3, 0).Unix()
var delta = max - min

func getDate() time.Time {
	sec := rand.Int63n(delta) + min
	return time.Unix(sec, 0)
}

func getNumber(num interface{}) interface{} {
	var value float64
	switch num.(type) {
	case int:
		return int(getRandomNumber(float64(num.(int))))
	case int8:
		return int8(getRandomNumber(float64(num.(int8))))
	case int32:
		return int32(getRandomNumber(float64(num.(int32))))
	case int64:
		return int64(getRandomNumber(float64(num.(int64))))
	case float32:
		return float32(getRandomNumber(float64(num.(float32))))
	case float64:
		return getRandomNumber(num.(float64))
	default:
		return value
	}
}

func getRandomNumber(x float64) float64 {
	mul := float64(1)
	for mul <= x {
		mul *= 10
	}
	v := mul * (rand.Float64() + .1)
	if x == math.Trunc(x) {
		return math.Round(v)
	}
	return math.Round(v*100) / 100
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
var fnames = []string{"Andrew", "Ava", "Becky", "Brian", "Cindy",
	"Connie", "David", "Dawn", "Elizabeth", "Emma",
	"Felix", "Frank", "George", "Grace", "Hector",
	"Henry", "Ian", "Isabella", "Jennifer", "John",
	"Kate", "Kenneth", "Linda", "Logan", "Mary",
	"Michael", "Nancy", "Noah", "Olivia", "Otis",
	"Patricia", "Peter", "Quentin", "Quinn", "Richard",
	"Robert", "Samuel", "Sophia", "Todd", "Tom",
	"Ulysses", "Umar", "Vincent", "Victoria", "Wesley",
	"Willaim", "Xavier", "Xena", "Yosef", "Yuri", "Zach", "Zoey",
}
var lnames = []string{"Smith", "Johnson", "Williams", "Brown", "Jones",
	"Miller", "Davis", "Garcia", "Rodriguez", "Chen",
	"Adams", "Arthur", "Bush", "Carter", "Clinton",
	"Eisenhower", "Ford", "Grant", "Harrison", "Hoover"}
