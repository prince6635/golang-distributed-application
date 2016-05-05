package main

import (
	"fmt"

	"github.com/golang-distributed-application/src/powerplant/coordinator"
)

func main() {
	coordinator.StartConsumingSensorData()

	var pause string
	fmt.Scanln(&pause)
}
