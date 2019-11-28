package main

import (
    "fmt"
    "log"
    "github.com/spf13/viper"
)

type clientInfo struct {
        Private_key         string
        Public_key          string
        Backup_dirs         []string
}

func main () {
	viper.SetConfigFile("config.yaml")
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
	fmt.Println(C.Private_key)
	fmt.Println(C.Public_key)
	fmt.Println(C.Backup_dirs)

}
