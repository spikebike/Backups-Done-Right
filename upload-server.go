package main

import (
	"./bdrservice" // defines BDR related protocols
	"./bdrsql"
	"./tlscon" // handles SSL connections
	"flag"
	"github.com/msbranco/goconfig"
	"log"
	"net"
)

var (
	configFile = flag.String("config", "../etc/config.cfg", "Defines where to load configuration from")
	newDB      = flag.Bool("new-db", false, "true = creates a new database | false = use existing database")
	debug_flag = flag.Bool("debug", false, "activates debug mode")
	debug      bool
)

type Request struct{}

func (Request) Request(in *bdrservice.RequestMessage, out *bdrservice.RequestACKMessage) error {
	var records int32

//	ack := &bdrservice.RequestACKMessage{}
	for _, blob := range in.Blobarray {
		log.Printf("server: blobarray=%v %T", *blob.Sha256, *blob.Sha256)
		records++
//		r := &bdrservice.RequestACKMessage{
//			RetVal: []bdrservice.RequestACKMessage_RetEnum{bdrservice.RequestACKMessage_HAVE},
//		r:=new(bdrservice.RequestACKMessage)
		out.RetVal=append(out.RetVal,records)
	}
//	*out.Received = ack
	return nil
}

func RequestFunc(conn net.Conn) {
	bdrservice.ServeRequestService(conn, Request{})
}

func main() {
	flag.Parse()
	log.Printf("loading config file from %s\n", *configFile)

	configF, err := goconfig.ReadConfigFile(*configFile)
	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}

	serverPort, err := configF.GetString("Server", "server_port")
	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}

	serverPrivKey, err := configF.GetString("Server", "private_key")
	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}

	serverPubKey, err := configF.GetString("Server", "public_key")
	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}

	dataBaseName, err := configF.GetString("Server", "sql_file")
	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}
	db, err := bdrsql.Init_db(dataBaseName, *newDB, debug)
	if err != nil {
		log.Printf("could not open %s, error: %s", dataBaseName, err)
	} else {
		log.Printf("opened database %v\n", dataBaseName)
	}
	err = bdrsql.CreateClientTables(db)
	if err != nil && debug == true {
		log.Printf("couldn't create tables: %s", err)
	} else {
		log.Printf("created tables\n")
	}

	fptr := RequestFunc
	tlscon.ServerTLSListen("0.0.0.0:"+serverPort, fptr, serverPrivKey, serverPubKey)

}
