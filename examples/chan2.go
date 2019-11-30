package main

import (
	"fmt"
)

var (
	dirChan = make(chan string,5)
	dirDone = make(chan bool,5)
)

func backupDir(dirChan chan string) {
   defer close(dirDone)

	for {
		select {
		case dir := <-dirChan:
			fmt.Printf("dir = %s\n", dir)
			if dir == "bar" {
				dirChan <- "frumple"
			}
		default:
			return
		}
	}
}

func main() {
	dirChan <- "foo"
	dirChan <- "bar"
	dirChan <- "baz"
	go backupDir(dirChan)
	<-dirDone
}
