// Copyright 2020 Kuei-chun Chen. All rights reserved.

package keyhole

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
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
	Logger                *mdb.Logger       `bson:"keyhole"`
	SourceStats           *mdb.ClusterStats `bson:"source"`
	TargetStats           *mdb.ClusterStats `bson:"target"`
	tlsCAFile             string
	tlsCertificateKeyFile string
	verbose               bool
}

// NewComparison returns *Comparison
func NewComparison(keyholeVersion string) *Comparison {
	comp := Comparison{Logger: mdb.NewLogger(keyholeVersion, "-compare")}
	comp.SourceStats = mdb.NewStats(keyholeVersion)
	comp.TargetStats = mdb.NewStats(keyholeVersion)
	return &comp
}

// SetTLSCAFile sets cloneDataOnly
func (p *Comparison) SetTLSCAFile(tlsCAFile string) {
	p.tlsCAFile = tlsCAFile
}

// SetTLSCertificateKeyFile sets cloneDataOnly
func (p *Comparison) SetTLSCertificateKeyFile(tlsCertificateKeyFile string) {
	p.tlsCertificateKeyFile = tlsCertificateKeyFile
}

// SetVerbose sets verbose
func (p *Comparison) SetVerbose(verbose bool) {
	p.verbose = verbose
}

// Run executes migration commands
func (p *Comparison) Run() error {
	var err error
	if len(flag.Args()) < 2 {
		return fmt.Errorf(`usage: keyhole -compare source_uri target_uri`)
	}
	if strings.HasSuffix(flag.Arg(0), "-stats.bson.gz") && strings.HasSuffix(flag.Arg(1), "-stats.bson.gz") {
		var data []byte
		var fd *bufio.Reader
		if fd, err = gox.NewFileReader(flag.Arg(0)); err != nil {
			return err
		}
		if data, err = ioutil.ReadAll(fd); err != nil {
			return err
		}
		if err = bson.Unmarshal(data, p.SourceStats); err != nil {
			panic(err)
			return err
		}
		if fd, err = gox.NewFileReader(flag.Arg(1)); err != nil {
			return err
		}
		if data, err = ioutil.ReadAll(fd); err != nil {
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
	if sourceConnString, err = mdb.ParseURI(flag.Arg(0)); err != nil {
		return err
	}
	sourceConnString.SSLCaFile = p.tlsCAFile
	sourceConnString.SSLClientCertificateKeyFile = p.tlsCertificateKeyFile
	if sourceClient, err = mdb.NewMongoClient(sourceConnString.String()); err != nil {
		return err
	}

	if targetConnString, err = mdb.ParseURI(flag.Arg(1)); err != nil {
		return err
	}
	// targetConnString.SSLCaFile = *targetTLSCAFile
	// targetConnString.SSLClientCertificateKeyFile = *targetTLSCertificateKeyFile
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
	for i, db := range p.TargetStats.Databases {
		dbMap[db.Name] = p.TargetStats.Databases[i]
	}
	// compare a few key metrics
	printer := message.NewPrinter(language.English)
	p.Logger.Log("\n=== Comparison Results (source vs. target) ===")
	p.Logger.Log(printer.Sprintf("Number of Databases:       \t%12d\t%12d", len(p.SourceStats.Databases), len(p.TargetStats.Databases)))
	for i, db := range p.SourceStats.Databases {
		collMap := map[string]mdb.Collection{}
		for i, coll := range dbMap[db.Name].Collections {
			collMap[coll.NS] = dbMap[db.Name].Collections[i]
		}
		p.Logger.Log(fmt.Sprintf("Database %v", db.Name))
		p.Logger.Log(printer.Sprintf(" ├─Number of Collections:\t%12d\t%12d", len(db.Collections), len(p.TargetStats.Databases[i].Collections)))
		p.Logger.Log(printer.Sprintf(" ├─Number of Indexes:    \t%12d\t%12d (all shards)", db.Stats.Indexes, dbMap[db.Name].Stats.Indexes))
		p.Logger.Log(printer.Sprintf(" ├─Number of Objects:    \t%12d\t%12d", db.Stats.Objects, dbMap[db.Name].Stats.Objects))
		p.Logger.Log(printer.Sprintf(" ├─Total Data Size:      \t%12s\t%12s", gox.GetStorageSize(db.Stats.DataSize), gox.GetStorageSize(dbMap[db.Name].Stats.DataSize)))
		p.Logger.Log(printer.Sprintf(" ├─Average Data Size:    \t%12s\t%12s", gox.GetStorageSize(db.Stats.AvgObjSize), gox.GetStorageSize(dbMap[db.Name].Stats.AvgObjSize)))
		p.Logger.Log(fmt.Sprintf(" └─Number of indexes"))
		for _, coll := range db.Collections {
			length := 0
			if val, ok := collMap[coll.NS]; ok {
				length = len(val.Indexes)
			}
			p.Logger.Log(fmt.Sprintf("   ├─%v:    \t%12d\t%12d", coll.NS, len(coll.Indexes), length))
		}
	}
	return err
}

// OutputBSON writes bson data to a file
func (p *Comparison) OutputBSON() error {
	if p.TargetStats.HostInfo.System.Hostname == "" {
		result := `Roles 'clusterMonitor' and 'readAnyDatabase' are required`
		return errors.New(result)
	}
	var err error
	var data []byte
	if data, err = bson.Marshal(p); err != nil {
		return err
	}
	outdir := "./out/"
	os.Mkdir(outdir, 0755)
	ofile := outdir + p.TargetStats.HostInfo.System.Hostname + "-compare.bson.gz"
	ofile = strings.ReplaceAll(ofile, ":", "_")
	if err = gox.OutputGzipped(data, ofile); err != nil {
		return err
	}
	p.Logger.Log(fmt.Sprintf(`bson data written to %v`, ofile))
	return err
}
