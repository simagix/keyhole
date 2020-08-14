// Copyright 2020 Kuei-chun Chen. All rights reserved.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/simagix/keyhole/mdb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

var version = "devel-xxxxxx"

func main() {
	var err error
	if version == "devel-xxxxxx" {
		version = "devel-" + time.Now().Format("20060102")
	}
	tlsCAFile := flag.String("tlsCAFile", "", "TLS CA file")
	tlsCertificateKeyFile := flag.String("tlsCertificateKeyFile", "", "TLS CertificateKey File")
	ver := flag.Bool("version", false, "print version number")
	flag.Parse()
	flagset := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { flagset[f.Name] = true })

	if *ver {
		fmt.Println("keyhole", version)
		return
	}

	if len(flag.Args()) == 0 {
		flag.PrintDefaults()
		fmt.Println("\nusage: keyhole [options] <connection_string>")
		os.Exit(0)
	}

	var uri = flag.Arg(0)
	var client *mongo.Client
	// connection string is required from here forward
	var connString connstring.ConnString
	if connString, err = mdb.ParseURI(uri); err != nil {
		log.Fatal(err)
	}
	uri = connString.String() // password can be injected if missing
	connString.SSLCaFile = *tlsCAFile
	connString.SSLClientCertificateKeyFile = *tlsCertificateKeyFile
	if client, err = mdb.NewMongoClient(uri, *tlsCAFile, *tlsCertificateKeyFile); err != nil {
		log.Fatal(err)
	}
	fmt.Println(mdb.GetClusterSummary(client))
}
