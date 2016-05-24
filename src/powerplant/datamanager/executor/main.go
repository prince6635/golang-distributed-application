package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"

	"github.com/golang-distributed-application/src/powerplant/datamanager"
	"github.com/golang-distributed-application/src/powerplant/dto"
	"github.com/golang-distributed-application/src/powerplant/queueutils"
)

// url for RabbitMQ channel, should be inside a config file
var url = "amqp://guest:guest@localhost:5672"

func main() {
	conn, ch := queueutils.GetChannel(url)
	defer conn.Close()
	defer ch.Close()

	msgs, err := ch.Consume(
		queueutils.PersistReadingsQueue, //queue string,
		"", //consumer string,
		// !!! need to verify the data has been saved to the database succesfully before ack
		false, //autoAck bool
		// !!! even it's running among multiple instances at the same time,
		//  this flag will fail the connect if it can't get an exclusive connection to this queue. (can also be used to keep things running in sequence.)
		true,  //exclusive bool,
		false, //noLocal bool,
		false, //noWait bool,
		nil)   //args amqp.Table)

	if err != nil {
		log.Fatalln("Failed to get access to messages.")
	}

	for msg := range msgs {
		// decode each message
		buffer := bytes.NewReader(msg.Body)
		decoder := gob.NewDecoder(buffer)
		sensorMsg := &dto.SensorMessage{}
		decoder.Decode(sensorMsg)

		err := datamanager.SaveReading(sensorMsg)
		if err != nil {
			fmt.Printf("reading data msg to be persisted: %+v", sensorMsg)
			log.Printf("Failed to save reading message to sensor %v. Error: %s", sensorMsg.Name, err.Error())
		} else {
			msg.Ack(false)
		}
	}
}
