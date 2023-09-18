// Copyright 2020-present Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/simagix/gox"
	"github.com/simagix/keyhole/mdb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// Comparison contains parameters of comparison parameters
type Comparison struct {
	Logger      *gox.Logger       `bson:"keyhole"`
	SourceStats *mdb.ClusterStats `bson:"source"`
	TargetStats *mdb.ClusterStats `bson:"target"`
	nocolor     bool
	verbose     bool
}

// NewComparison returns *Comparison
func NewComparison(keyholeVersion string) *Comparison {
	comp := Comparison{Logger: gox.GetLogger(keyholeVersion)}
	comp.SourceStats = mdb.NewClusterStats(keyholeVersion)
	comp.TargetStats = mdb.NewClusterStats(keyholeVersion)
	return &comp
}

// SetNoColor set nocolor flag
func (p *Comparison) SetNoColor(nocolor bool) {
	p.nocolor = nocolor
}

// SetVerbose sets verbose
func (p *Comparison) SetVerbose(verbose bool) {
	p.verbose = verbose
}

// Compare executes compare commands
func (p *Comparison) Compare(source string, target string) error {
	var err error
	if strings.HasSuffix(source, "-stats.bson.gz") && strings.HasSuffix(target, "-stats.bson.gz") { // compare files
		var data []byte
		var fd *bufio.Reader
		if fd, err = gox.NewFileReader(source); err != nil {
			return err
		}
		if data, err = io.ReadAll(fd); err != nil {
			return err
		}
		if err = bson.Unmarshal(data, p.SourceStats); err != nil {
			return err
		}
		if fd, err = gox.NewFileReader(target); err != nil {
			return err
		}
		if data, err = io.ReadAll(fd); err != nil {
			return err
		}
		if err = bson.Unmarshal(data, p.TargetStats); err != nil {
			return err
		}
		return p.compare()
	}
	var sourceClient *mongo.Client
	var targetClient *mongo.Client
	var sourceConnString connstring.ConnString
	var targetConnString connstring.ConnString
	// connection string is required from here forward
	if sourceConnString, err = mdb.ParseURI(source); err != nil {
		return err
	}
	if sourceClient, err = mdb.NewMongoClient(sourceConnString.String()); err != nil {
		return err
	}

	if targetConnString, err = mdb.ParseURI(target); err != nil {
		return err
	}
	if targetClient, err = mdb.NewMongoClient(targetConnString.String()); err != nil {
		return err
	}
	var wg = gox.NewWaitGroup(1)
	wg.Add(1)
	go func(stats *mdb.ClusterStats, client *mongo.Client, connString connstring.ConnString) {
		defer wg.Done()
		stats.SetVerbose(false)
		if err = stats.GetClusterStats(client, connString); err != nil {
			result := `Roles 'clusterMonitor' and 'readAnyDatabase' are required`
			log.Fatal(result)
		}
	}(p.SourceStats, sourceClient, sourceConnString)
	wg.Add(1)
	go func(stats *mdb.ClusterStats, client *mongo.Client, connString connstring.ConnString) {
		defer wg.Done()
		stats.SetVerbose(false)
		if err = stats.GetClusterStats(client, connString); err != nil {
			result := `Roles 'clusterMonitor' and 'readAnyDatabase' are required`
			log.Fatal(result)
		}
	}(p.TargetStats, targetClient, targetConnString)
	wg.Wait()
	return p.compare()
}

func (p *Comparison) compare() error {
	var err error
	// build target stats map
	dbMap := map[string]mdb.Database{}
	for i, db := range *p.TargetStats.Databases {
		dbMap[db.Name] = (*p.TargetStats.Databases)[i]
	}
	// compare a few key metrics
	codeDefault := mdb.CodeDefault
	if p.nocolor {
		codeDefault = ""
	}
	printer := message.NewPrinter(language.English)
	p.Logger.Info("=== Comparison Results (source vs. target) ===")
	p.Logger.Info(printer.Sprintf("Number of Databases:       \t%12d%v\t%12d%v",
		len(*p.SourceStats.Databases), p.getColor(int64(len(*p.SourceStats.Databases)), int64(len(*p.TargetStats.Databases))), len(*p.TargetStats.Databases), codeDefault))
	for _, db := range *p.SourceStats.Databases {
		collMap := map[string]mdb.Collection{}
		for i, coll := range dbMap[db.Name].Collections {
			collMap[coll.NS] = dbMap[db.Name].Collections[i]
		}
		p.Logger.Info(fmt.Sprintf("Database %v", db.Name))
		nColl := len(dbMap[db.Name].Collections)
		p.Logger.Info(printer.Sprintf(" ├─Number of Collections:\t%12d%v\t%12d%v",
			len(db.Collections), p.getColor(int64(len(db.Collections)), int64(nColl)), nColl, codeDefault))
		p.Logger.Info(printer.Sprintf(" ├─Number of Indexes:    \t%12d%v\t%12d%v (all shards)",
			db.Stats.Indexes, p.getColor(db.Stats.Indexes, dbMap[db.Name].Stats.Indexes), dbMap[db.Name].Stats.Indexes, codeDefault))
		p.Logger.Info(printer.Sprintf(" ├─Number of Objects:    \t%12d%v\t%12d%v",
			db.Stats.Objects, p.getColor(db.Stats.Objects, dbMap[db.Name].Stats.Objects), dbMap[db.Name].Stats.Objects, codeDefault))
		p.Logger.Info(printer.Sprintf(" ├─Total Data Size:      \t%12s%v\t%12s%v",
			gox.GetStorageSize(db.Stats.DataSize), p.getColor(db.Stats.DataSize, dbMap[db.Name].Stats.DataSize), gox.GetStorageSize(dbMap[db.Name].Stats.DataSize), codeDefault))
		p.Logger.Info(printer.Sprintf(" ├─Average Data Size:    \t%12s%v\t%12s%v",
			gox.GetStorageSize(db.Stats.AvgObjSize), p.getColor(db.Stats.AvgObjSize, dbMap[db.Name].Stats.AvgObjSize), gox.GetStorageSize(dbMap[db.Name].Stats.AvgObjSize), codeDefault))
		p.Logger.Info(" └─Number of indexes:")
		for _, coll := range db.Collections {
			length := 0
			if val, ok := collMap[coll.NS]; ok {
				length = len(val.Indexes)
			}
			p.Logger.Info(fmt.Sprintf("   ├─%v:    \t%12d\t%12d", coll.NS, len(coll.Indexes), length))
		}
	}
	return err
}

func (p *Comparison) getColor(a int64, b int64) string {
	if p.nocolor {
		if a != b {
			return " ≠"
		}
		return ""
	}
	if a != b {
		return mdb.CodeRed
	}
	return mdb.CodeDefault
}

// OutputBSON writes bson data to a file
func (p *Comparison) OutputBSON() error {
	if p.TargetStats.HostInfo.System.Hostname == "" {
		result := `roles 'clusterMonitor' and 'readAnyDatabase' are required`
		return errors.New(result)
	}
	var err error
	var data []byte
	if data, err = bson.Marshal(p); err != nil {
		return err
	}
	os.Mkdir(outdir, 0755)
	basename := p.TargetStats.HostInfo.System.Hostname
	basename = strings.ReplaceAll(basename, ":", "_")
	ofile := fmt.Sprintf(`%v/%v-compare.bson.gz`, outdir, basename)
	i := 1
	for mdb.DoesFileExist(ofile) {
		ofile = fmt.Sprintf(`%v/%v.%d-compare.bson.gz`, outdir, basename, i)
		i++
	}
	if err = gox.OutputGzipped(data, ofile); err != nil {
		return err
	}
	p.Logger.Info(fmt.Sprintf(`bson data written to %v`, ofile))
	return err
}
