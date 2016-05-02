package main

import "github.com/golang-distributed-application/src/powerplant/sensors"

func main() {
	// test.PrintDummy()
	// test.RabbitMQ()
	sensor.StartPublishingSensorData()
}
