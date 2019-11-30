package main

import (
	"fmt"
)

var (
	dirChan = make(chan string,5)
	dirDone = make(chan bool,5)
)

func backupDir(dirChan chan string) {
LOOP:
	for {
		select {
		case dir := <-dirChan:
			fmt.Printf("dir = %s\n", dir)
			if dir == "bar" {
				dirChan <- "frumple"
			}
		default:
			break LOOP
		}
	}
	dirDone <- true
}

func main() {
	dirChan <- "foo"
	dirChan <- "bar"
	dirChan <- "baz"
	go backupDir(dirChan)
	<-dirDone
}
