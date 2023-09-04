// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/simagix/gox"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// IndexStats holder indexes reader struct
type IndexStats struct {
	Databases []Database  `bson:"databases"`
	Logger    *gox.Logger `bson:"keyhole"`

	fastMode bool
	filename string
	nocolor  bool
	verbose  bool
	version  string
}

// Accesses stores index accesses
type Accesses struct {
	Ops   int       `json:"ops" bson:"ops"`
	Since time.Time `json:"since" bson:"since"`
}

// IndexUsage stores index accesses
type IndexUsage struct {
	Accesses Accesses `json:"accesses" bson:"accesses"`
	Host     string   `json:"host" bson:"host"`
	Name     string   `json:"name" bson:"name"`
	Shard    string   `json:"shard" bson:"shard"`
}

// Index stores indexes stats
type Index struct {
	Background              bool   `json:"background" bson:"background"`
	Collation               bson.D `json:"collation" bson:"collation"`
	ExpireAfterSeconds      int32  `json:"expireAfterSeconds" bson:"expireAfterSeconds,truncate,omitempty"`
	Key                     bson.D `json:"key" bson:"key"`
	Name                    string `json:"name" bson:"name,truncate"`
	PartialFilterExpression bson.D `json:"partialFilterExpression" bson:"partialFilterExpression"`
	Sparse                  bool   `json:"sparse" bson:"sparse"`
	Unique                  bool   `json:"unique" bson:"unique"`
	Version                 int32  `json:"v" bson:"v,truncate"`
	Weights                 bson.D `json:"weights" bson:"weights"`

	EffectiveKey string       `json:"effectiveKey" bson:"effectiveKey"`
	Fields       []string     `json:"fields" bson:"fields"`
	IsDupped     bool         `json:"isDupped" bson:"isDupped"`
	IsShardKey   bool         `json:"isShardkey" bson:"isShardkey"`
	KeyString    string       `json:"keyString" bson:"keyString"`
	TotalOps     int          `json:"totalOps" bson:"totalOps"`
	Usage        []IndexUsage `json:"usage" bson:"usage"`
}

const (
	indexExt = "-index.bson.gz"
)

// NewIndexStats establish seeding parameters
func NewIndexStats(version string) *IndexStats {
	hostname, _ := os.Hostname()
	return &IndexStats{version: version, Logger: gox.GetLogger(version),
		filename: hostname + indexExt, Databases: []Database{}}
}

// SetFastMode sets fastMode mode
func (ix *IndexStats) SetFastMode(fastMode bool) {
	ix.fastMode = fastMode
}

// SetFilename sets output file name
func (ix *IndexStats) SetFilename(filename string) {
	ix.filename = strings.Replace(filename, ":", "_", -1)
}

// SetClusterDetailsFromFile File sets cluster details from a file
func (ix *IndexStats) SetClusterDetailsFromFile(filename string) error {
	if !strings.HasSuffix(filename, indexExt) && !strings.HasSuffix(filename, "-stats.bson.gz") {
		return errors.New("unsupported file type")
	}
	var data []byte
	var err error
	var fd *bufio.Reader
	if fd, err = gox.NewFileReader(filename); err != nil {
		return err
	}
	if data, err = io.ReadAll(fd); err != nil {
		return err
	}
	return bson.Unmarshal(data, &ix)
}

// SetNoColor set nocolor flag
func (ix *IndexStats) SetNoColor(nocolor bool) {
	ix.nocolor = nocolor
}

// SetVerbose sets verbose level
func (ix *IndexStats) SetVerbose(verbose bool) {
	ix.verbose = verbose
	if verbose && ix.Logger != nil {
		ix.Logger.SetLoggerLevel(gox.Debug)
	}
}

// GetIndexes list all indexes of collections of databases
func (ix *IndexStats) GetIndexes(client *mongo.Client) ([]Database, error) {
	var err error
	var dbNames []string
	var collections []Collection
	ix.Databases = []Database{}
	var databases []Database
	if dbNames, err = GetDatabaseNames(client); err != nil {
		return databases, err
	}
	cnt := 0
	for _, name := range dbNames {
		if name == "admin" || name == "config" || name == "local" {
			ix.Logger.Debug("skip ", name)
			continue
		}
		cnt++
		ix.Logger.Debug("checking ", name)
		if collections, err = ix.GetIndexesFromDB(client, name); err != nil {
			return ix.Databases, err
		}
		ix.Databases = append(ix.Databases, Database{Name: name, Collections: collections})
	}
	if cnt == 0 && ix.verbose {
		ix.Logger.Info("No database is available")
	}
	ix.Logger.Info(`GetIndexes ends`)
	return ix.Databases, err
}

// GetIndexesFromDB list all indexes of collections of a database
func (ix *IndexStats) GetIndexesFromDB(client *mongo.Client, db string) ([]Collection, error) {
	var err error
	var cur *mongo.Cursor
	var ctx = context.Background()
	var collections []Collection
	ix.Logger.Debugf(`GetIndexesFromDB(%v)`, db)
	if cur, err = client.Database(db).ListCollections(ctx, bson.M{}); err != nil {
		return collections, err
	}
	defer cur.Close(ctx)
	collectionNames := []string{}
	for cur.Next(ctx) {
		var elem struct {
			Name string `bson:"name"`
			Type string `bson:"type"`
		}
		if err = cur.Decode(&elem); err != nil {
			continue
		}
		if strings.HasPrefix(elem.Name, "system.") || (elem.Type != "" && elem.Type != "collection") {
			ix.Logger.Debug("skip ", elem.Name)
			continue
		}
		collectionNames = append(collectionNames, elem.Name)
	}

	sort.Strings(collectionNames)
	for _, v := range collectionNames {
		var collection = Collection{NS: db + "." + v, Name: v}
		if collection.Indexes, err = ix.GetIndexesFromCollection(client, client.Database(db).Collection(v)); err != nil {
			return collections, err
		}
		collections = append(collections, collection)
	}
	return collections, nil
}

// GetIndexesFromCollection gets indexes from a collection
func (ix *IndexStats) GetIndexesFromCollection(client *mongo.Client, collection *mongo.Collection) ([]Index, error) {
	var err error
	var ctx = context.Background()
	var pipeline = MongoPipeline(`{"$indexStats": {}}`)
	var list []Index
	var icur *mongo.Cursor
	var scur *mongo.Cursor
	db := collection.Database().Name()
	ix.Logger.Debugf(`GetIndexesFromCollection from %v.%v`, db, collection.Name())

	if strings.HasPrefix(collection.Name(), "system.") {
		ix.Logger.Debug("skip ", collection.Name())
		return list, nil
	}

	var indexStats = []IndexUsage{}
	if scur, err = collection.Aggregate(ctx, pipeline); err != nil {
		ix.Logger.Error(err)
	} else {
		for scur.Next(ctx) {
			var result IndexUsage
			if err = scur.Decode(&result); err != nil {
				ix.Logger.Error(err)
				continue
			}
			indexStats = append(indexStats, result)
		}
		scur.Close(ctx)
	}

	cmd := bson.D{{Key: "listIndexes", Value: collection.Name()}}
	if icur, err = client.Database(db).RunCommandCursor(ctx, cmd); err != nil {
		ix.Logger.Error(err)
		return list, err
	}
	defer icur.Close(ctx)
	indexesFound := map[int]bool{}

	for icur.Next(ctx) {
		o := Index{ExpireAfterSeconds: -1}
		if err = icur.Decode(&o); err != nil {
			ix.Logger.Error(err)
			continue
		}
		var strbuf bytes.Buffer
		fields := []string{}
		for n, value := range o.Key {
			fields = append(fields, value.Key)
			if n == 0 {
				strbuf.WriteString("{ ")
			}
			strbuf.WriteString(value.Key + ": " + fmt.Sprint(value.Value))
			if n == len(o.Key)-1 {
				strbuf.WriteString(" }")
			} else {
				strbuf.WriteString(", ")
			}
		}
		o.Fields = fields
		o.KeyString = strbuf.String()
		// Check shard keys
		var v map[string]interface{}
		ns := collection.Database().Name() + "." + collection.Name()
		ix.Logger.Debug("GetIndexesFromCollection ", ns, o.KeyString)
		if !ix.fastMode {
			if err = client.Database("config").Collection("collections").FindOne(ctx, bson.M{"_id": ns, "key": o.Key}).Decode(&v); err == nil {
				o.IsShardKey = true
			}
		}
		o.EffectiveKey = strings.Replace(o.KeyString[2:len(o.KeyString)-2], ": -1", ": 1", -1)
		o.Usage = []IndexUsage{}
		for i, result := range indexStats {
			if result.Name == o.Name {
				indexesFound[i] = true
				o.TotalOps += result.Accesses.Ops
				o.Usage = append(o.Usage, result)
			}
		}
		list = append(list, o)
	}
	sort.Slice(list, func(i, j int) bool { return (list[i].EffectiveKey < list[j].EffectiveKey) })
	for i, o := range list {
		if o.KeyString != "{ _id: 1 }" && !o.IsShardKey {
			list[i].IsDupped = checkIfDupped(o, list)
		}
	}
	if len(indexesFound) != len(indexStats) {
		for i := 0; i < len(indexStats); i++ {
			if !indexesFound[i] {
				ns := collection.Database().Name() + "." + collection.Name()
				ix.Logger.Warnf(`inconsistent index '%v' of namespace '%v' on shard '%v'`,
					indexStats[i].Name, ns, indexStats[i].Shard)
			}
		}
	}
	return list, nil
}

// check if an index is a dup of others
func checkIfDupped(doc Index, list []Index) bool {
	if strings.Index(doc.KeyString, "2dsphere") > 0 {
		return false
	}
	for _, o := range list {
		if strings.Index(o.KeyString, "2dsphere") > 0 {
			continue
		} else if !o.IsDupped && doc.Fields[0] == o.Fields[0] && doc.KeyString != o.KeyString && len(o.Fields) >= len(doc.Fields) {
			// check indexes if not marked as dupped, has the same first field, and more or equal number of fields
			nmatched := 0
			for i, fld := range doc.Fields {
				if i == 0 {
					continue
				}
				for j, field := range o.Fields {
					if j > 0 && fld == field {
						nmatched++
						break
					}
				}
			}
			if nmatched == len(doc.Fields)-1 {
				return true
			}
		}
	}
	return false
}

// OutputBSON writes index stats bson to a file
func (ix *IndexStats) OutputBSON() (string, []byte, error) {
	var err error
	var bsond bson.D
	var data []byte
	var ofile string
	if data, err = bson.Marshal(ix); err != nil {
		return ofile, data, err
	}
	bson.Unmarshal(data, &bsond)
	if data, err = bson.Marshal(bsond); err != nil {
		return ofile, data, err
	}

	os.Mkdir(outdir, 0755)
	idx := strings.Index(ix.filename, indexExt)
	basename := ix.filename[:idx]
	ofile = fmt.Sprintf(`%v/%v%v`, outdir, basename, indexExt)
	i := 1
	for DoesFileExist(ofile) {
		ofile = fmt.Sprintf(`%v/%v.%d%v`, outdir, basename, i, indexExt)
		i++
	}

	if err = gox.OutputGzipped(data, ofile); err == nil {
		fmt.Println("Index stats is written to", ofile)
	}
	return ofile, data, err
}

// OutputJSON writes json data to a file
func (ix *IndexStats) OutputJSON() error {
	var err error
	var data []byte
	if data, err = bson.MarshalExtJSON(ix, false, false); err != nil {
		return err
	}
	os.Mkdir(outdir, 0755)
	ofile := fmt.Sprintf("%v/%v", outdir, strings.ReplaceAll(filepath.Base(ix.filename), "bson.gz", "json"))
	os.WriteFile(ofile, data, 0644)
	fmt.Println("json data written to", ofile)
	return err
}

// Print prints indexes
func (ix *IndexStats) Print() {
	ix.PrintIndexesOf(ix.Databases)
}

// PrintIndexesOf prints indexes
func (ix *IndexStats) PrintIndexesOf(databases []Database) {
	for _, db := range databases {
		for _, coll := range db.Collections {
			var buffer bytes.Buffer
			ns := coll.NS
			buffer.WriteString("\n")
			buffer.WriteString(ns)
			buffer.WriteString(":\n")
			for _, o := range coll.Indexes {
				font := CodeDefault
				tailCode := CodeDefault
				if ix.nocolor {
					font = ""
					tailCode = ""
				}
				if o.KeyString == "{ _id: 1 }" {
					buffer.WriteString(fmt.Sprintf("%v  %v%v", font, o.KeyString, tailCode))
				} else if o.IsShardKey {
					buffer.WriteString(fmt.Sprintf("%v* %v%v", font, o.KeyString, tailCode))
				} else if o.IsDupped {
					if !ix.nocolor {
						font = CodeRed
					}
					buffer.WriteString(fmt.Sprintf("%vx %v%v", font, o.KeyString, tailCode))
				} else if o.TotalOps == 0 && o.ExpireAfterSeconds < 0 {
					if !ix.nocolor {
						font = CodeBlue
					}
					buffer.WriteString(fmt.Sprintf("%v? %v%v", font, o.KeyString, tailCode))
				} else {
					buffer.WriteString(fmt.Sprintf("  %v", o.KeyString))
				}

				for _, u := range o.Usage {
					buffer.Write([]byte("\n\thost: " + u.Host + ", ops: " + fmt.Sprintf("%v", u.Accesses.Ops) + ", since: " + fmt.Sprintf("%v", u.Accesses.Since)))
				}
				buffer.WriteString("\n")
			}
			fmt.Println(buffer.String())
		}
	}
}

// IndexNS defines from and to namespaces
type IndexNS struct {
	From string
	To   string
}

// CreateIndexes creates indexes
func (ix *IndexStats) CreateIndexes(client *mongo.Client, namespaces ...[]string) error {
	indexNamespaces := []IndexNS{}
	if len(namespaces) > 0 {
		for _, v := range namespaces[0] {
			indexNamespaces = append(indexNamespaces, IndexNS{From: v, To: v})
		}
	}
	return ix.CopyIndexesWithDest(client, indexNamespaces, true)
}

// CopyIndexes copies indexes
func (ix *IndexStats) CopyIndexes(client *mongo.Client, isDrop bool, namespaces ...[]string) error {
	indexNamespaces := []IndexNS{}
	if len(namespaces) > 0 {
		for _, v := range namespaces[0] {
			indexNamespaces = append(indexNamespaces, IndexNS{From: v, To: v})
		}
	}
	return ix.CopyIndexesWithDest(client, indexNamespaces, isDrop)
}

// CopyIndexesWithDest copies indexes
func (ix *IndexStats) CopyIndexesWithDest(client *mongo.Client, namespaces []IndexNS, isDrop bool) error {
	var ctx = context.Background()
	var err error
	namespaceMap := map[string]bool{}
	indexMap := map[string]string{}
	if len(namespaces) > 0 {
		for _, ns := range namespaces {
			namespaceMap[ns.From] = true
			indexMap[ns.From] = ns.To
		}
	}
	for _, db := range ix.Databases {
		for _, coll := range db.Collections {
			dbName := db.Name
			collName := coll.Name
			ns := dbName + "." + collName
			if SkipNamespace(ns, namespaceMap) {
				continue
			}
			if indexMap[ns] != "" {
				dbName, collName = SplitNamespace(indexMap[ns])
				ns = dbName + "." + collName
			}
			if isDrop {
				var doc bson.M
				cmd := bson.D{{Key: "dropIndexes", Value: collName}, {Key: "index", Value: "*"}}
				client.Database(dbName).RunCommand(ctx, cmd).Decode(&doc)
			}
			collection := client.Database(dbName).Collection(collName)
			indexes := []mongo.IndexModel{}
			for _, o := range coll.Indexes {
				opt := options.Index()
				// opt.SetVersion(o.Version)
				opt.SetName(o.Name)
				if o.Background {
					opt.SetBackground(o.Background)
				}
				if o.ExpireAfterSeconds >= 0 {
					opt.SetExpireAfterSeconds(o.ExpireAfterSeconds)
				}
				if o.Unique {
					opt.SetUnique(o.Unique)
				}
				if o.Sparse {
					opt.SetSparse(o.Sparse)
				}
				var collation *options.Collation
				if o.Collation != nil {
					data, err := bson.Marshal(o.Collation)
					if err != nil {
						return err
					}
					bson.Unmarshal(data, &collation)
					opt.SetCollation(collation)
				}
				if o.PartialFilterExpression != nil {
					opt.SetPartialFilterExpression(o.PartialFilterExpression)
				}
				if o.Weights != nil {
					opt.SetWeights(o.Weights)
				}
				if ix.verbose {
					ix.Logger.Info(fmt.Sprintf(`creating index %v on %v `, o.KeyString, ns))
				}
				indexes = append(indexes, mongo.IndexModel{Keys: o.Key, Options: opt})
				if o.Key.Map()["_id"] != nil {
					collNames, _ := client.Database(dbName).ListCollectionNames(ctx, bson.D{})
					var exists bool
					for _, name := range collNames {
						if name == coll.Name {
							exists = true
							break
						}
					}
					if !exists {
						collOpts := options.CreateCollection()
						if collation != nil {
							collOpts.SetCollation(collation)
						}
						if err = client.Database(dbName).CreateCollection(ctx, collName, collOpts); err != nil {
							return err
						}
						if _, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: o.Key, Options: opt}); err != nil {
							return err
						}
					}
				}
			}
			if _, err = collection.Indexes().CreateMany(ctx, indexes); err != nil {
				return err
			}
		}
	}
	return err
}

// GetDatabaseNames gets all database names
func GetDatabaseNames(client *mongo.Client) ([]string, error) {
	var err error
	var names []string
	var result mongo.ListDatabasesResult
	if result, err = client.ListDatabases(context.Background(), bson.M{}); err != nil {
		return names, err
	}
	for _, db := range result.Databases {
		names = append(names, db.Name)
	}
	return names, err
}
