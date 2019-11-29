package main

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
)

type clientInfo struct {
	PrivateKey   string
	PublicKey    string
	BackupDirs   []string
	ExcludeDirs  []string
	Threads      int
	SqlFile      string
	QueueBlobDir string
}

func main() {
	viper.SetConfigFile("etc/config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
	fmt.Printf("Ok\n")
	prod := viper.Sub("client")
	var C clientInfo
	err := prod.Unmarshal(&C)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}
	fmt.Println(C.PrivateKey)
	fmt.Println(C.PublicKey)
	fmt.Println(C.BackupDirs)
   log.Printf("backing up these directories: %s\n", C.BackupDirs)
	fmt.Println(C.Threads)
	fmt.Println(C.SqlFile)
	fmt.Println(C.ExcludeDirs)
	fmt.Println(C.QueueBlobDir)
}
