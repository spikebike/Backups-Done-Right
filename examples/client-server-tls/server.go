package main

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"log"
	"net"
)

func main() {
	cert, err := tls.LoadX509KeyPair("certs/server.pem", "certs/server.key")
	if err != nil {
		log.Fatalf("server: loadkeys: %s", err)
	}
	config := tls.Config{Certificates: []tls.Certificate{cert}, ClientAuth: tls.RequireAnyClientCert}
	config.Rand = rand.Reader
	service := "0.0.0.0:8000"
	listener, err := tls.Listen("tcp", service, &config)
	if err != nil {
		log.Fatalf("server: listen: %s", err)
	}
	log.Print("server: listening")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("server: accept: %s", err)
			break
		}
		log.Printf("server: accepted from %s", conn.RemoteAddr())
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	tlscon, ok := conn.(*tls.Conn)
	if ok {
		log.Print("server: conn: type assert to TLS succeedded")
		err := tlscon.Handshake()
		if err != nil {
			log.Fatalf("server: handshake failed: %s", err)
		} else {
			log.Print("server: conn: Handshake completed")
		}
		state := tlscon.ConnectionState()
		log.Println("Server: client public key is:")
		for _, v := range state.PeerCertificates {
			log.Print(x509.MarshalPKIXPublicKey(v.PublicKey))
		}
		buf := make([]byte, 512)
		for {
			log.Print("server: conn: waiting")
			n, err := conn.Read(buf)
			if err != nil {
				if err != nil {
					log.Printf("server: conn: read: %s", err)
				}
				break

			}
			log.Printf("server: conn: echo %q\n", string(buf[:n]))
			n, err = conn.Write(buf[:n])
			log.Printf("server: conn: wrote %d bytes", n)
			if err != nil {
				log.Printf("server: write: %s", err)
				break
			}
		}
	}
	log.Println("server: conn: closed")
}
