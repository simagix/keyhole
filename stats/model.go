package stats

import (
	"fmt"
	"math/rand"
)

var x = 1024

// DBName -
var DBName = fmt.Sprintf("_KEYHOLE_%X", x+x*rand.Intn(x))

// Model -
type Model struct {
	ID          string `json:"_id" bson:"_id"`
	Name        string
	Description string
	Year        int
}
