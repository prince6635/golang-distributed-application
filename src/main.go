package main

import (
	"fmt"

	"github.com/golang-distributed-application/src/test"
)

func main() {
	test.PrintDummy()

	// test RabbitMQ
	go test.Publish()
	go test.Consume()

	// keep goroutines running
	var hang string
	fmt.Scanln(&hang)
}
