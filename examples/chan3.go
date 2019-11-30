package main

import (
	"fmt"
)

var (
   numChan = make(chan int64,5)
   replyChan = make(chan int64,5)
)

func add(numChan chan int64,replyChan chan int64) {
	var j int64
	for i := range numChan {
		fmt.Printf("got %d\n",i)
		j=i+2;
		replyChan <- j
	}
}

func main() {
	numChan <-2
	go add(numChan,replyChan)
	fmt.Printf("answer is %d\n",<-replyChan)
}
