// from https://stackoverflow.com/questions/14404757/how-to-encrypt-and-decrypt-plain-text-with-a-rsa-keys-in-go // Kyle Lemons
package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"io/ioutil"
	"log"
)

// Command-line flags
var (
	keyFile = flag.String("key", "id_rsa", "Path to RSA private key")
	inFile  = flag.String("in", "in.txt", "Path to input file")
	outFile = flag.String("out", "out.txt", "Path to output file")
	label   = flag.String("label", "", "Label to use (filename by default)")
	decrypt = flag.Bool("decrypt", false, "Decrypt instead of encrypting")
)

func main() {
	flag.Parse()

	// Read the input file
	in, err := ioutil.ReadFile(*inFile)
	if err != nil {
		log.Fatalf("input file: %s", err)
	}

	// Read the private key
	pemData, err := ioutil.ReadFile(*keyFile)
	if err != nil {
		log.Fatalf("read key file: %s", err)
	}

	// Extract the PEM-encoded data block
	block, _ := pem.Decode(pemData)
	if block == nil {
		log.Fatalf("bad key data: %s", "not PEM-encoded")
	}
	if got, want := block.Type, "RSA PRIVATE KEY"; got != want {
		log.Fatalf("unknown key type %q, want %q", got, want)
	}

	// Decode the RSA private key
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		log.Fatalf("bad private key: %s", err)
	}

	var out []byte
	if *decrypt {
		if *label == "" {
			*label = *outFile
		}
		// Decrypt the data
		out, err = rsa.DecryptOAEP(sha1.New(), rand.Reader, priv, in, []byte(*label))
		if err != nil {
			log.Fatalf("decrypt: %s", err)
		}
	} else {
		if *label == "" {
			*label = *inFile
		}
		out, err = rsa.EncryptOAEP(sha1.New(), rand.Reader, &priv.PublicKey, in, []byte(*label))
		if err != nil {
			log.Fatalf("encrypt: %s", err)
		}
	}

	// Write data to output file
	if err := ioutil.WriteFile(*outFile, out, 0600); err != nil {
		log.Fatalf("write output: %s", err)
	}
}
