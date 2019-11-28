package main

import (
	"./bdrservice" // defines BDR related protocols
	"./tlscon"     // handles SSL connections
	"crypto/rand"
	"crypto/sha256"
	"flag"
	"fmt"
	"github.com/msbranco/goconfig"
	"log"
)

var (
	configFile = flag.String("config", "../etc/config.cfg", "Defines where to load configuration from")
	newDB      = flag.Bool("new-db", false, "true = creates a new database | false = use existing database")
	debug_flag = flag.Bool("debug", false, "activates debug mode")
	debug      bool
)

func main() {

	flag.Parse()
	log.Printf("loading config file from %s\n", *configFile)

	configF, err := goconfig.ReadConfigFile(*configFile)
	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}

	clientPrivKey, err := configF.GetString("Client", "private_key")
	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}

	clientPubKey, err := configF.GetString("Client", "public_key")
	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}

	server, err := configF.GetString("Client", "server")
	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}
	serverPort, err := configF.GetString("Client", "server_port")
	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}

	// connect to server
	conn, err := tlscon.OpenTLSClient(server+":"+serverPort, clientPrivKey, clientPubKey)
	if err != nil {
		log.Fatalf("dial: %s", err)
	}

	request := bdrservice.NewRequestServiceClient(conn)
	req := &bdrservice.RequestMessage{}
	ack := &bdrservice.RequestACKMessage{}
	for i := 0; i < 4; i++ {
		// read in 16 bytes from /dev/urandom to sha256
		randBytes := make([]byte, 16)
		rand.Read(randBytes)

		// get its size
		size := int32(len(randBytes))

		// create a new hash, and do a crypto hash of the random bytes.
		sha := sha256.New()
		sha.Write(randBytes)
		strhash := fmt.Sprintf("%x", sha.Sum(nil))
		fmt.Printf("i=%d sha=%s size=%d\n", i, strhash, size)

		req.Blobarray = append(req.Blobarray, &bdrservice.RequestMessageBlob{Sha256: &strhash, Bsize: &size})
	}
	if err := request.Request(req, ack); err != nil {
		log.Fatalf("Add failed with: %s", err)
	}
	fmt.Printf("Server ACKs %d records\n\n", ack.RetVal)
}
