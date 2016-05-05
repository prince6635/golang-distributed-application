package coordinator

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/golang-distributed-application/src/powerplant/dto"
	"github.com/golang-distributed-application/src/powerplant/queueutils"
	"github.com/streadway/amqp"
)

/*
1, discover all the sensor data queues
2, for each sensor data queue, consume/receive its messages
3, transalate the messages into events and send to event aggregator
*/

const url = "amqp://guest:guest@localhost:5672"

type QueuesListener struct {
	conn    *amqp.Connection
	ch      *amqp.Channel
	sources map[string]<-chan amqp.Delivery // message receiving channel
}

func NewQueuesListener() *QueuesListener {
	ql := QueuesListener{
		sources: make(map[string]<-chan amqp.Delivery),
	}

	ql.conn, ql.ch = queueutils.GetChannel(url)
	return &ql
}

func StartConsumingSensorData() {
	ql := NewQueuesListener()
	go ql.ListenForNewSource()
}

func (ql *QueuesListener) ListenForNewSource() {
	q := queueutils.GetQueue("", ql.ch)

	// bind the queue to receive fan-out messages
	ql.ch.QueueBind(
		q.Name,       //name string,
		"",           //key string,
		"amq.fanout", //exchange string,
		false,        //noWait bool,
		nil)          //args amqp.Table)

	// the following each message means a new sensor is coming online
	msgs, _ := ql.ch.Consume(
		q.Name, //queue string,
		"",     //consumer string,
		true,   //autoAck bool,
		false,  //exclusive bool,
		false,  //noLocal bool,
		false,  //noWait bool,
		nil)    //args amqp.Table)

	fmt.Println("Listening for new sources")
	for msg := range msgs {
		fmt.Println("New source discovered")

		// for this new source (sensor data queue), start to receive its reading data
		sensorDataQueueName := string(msg.Body)
		sourceChann, _ := ql.ch.Consume(
			sensorDataQueueName, //queue string, sensor data queue's name,
			"",                  //consumer string,
			true,                //autoAck bool,
			false,               //exclusive bool,
			false,               //noLocal bool,
			false,               //noWait bool,
			nil)                 //args amqp.Table)

		// add the new source to a map for de-dup, also call AddListener for consuming sensor reading data
		if ql.sources[sensorDataQueueName] == nil {
			ql.sources[sensorDataQueueName] = sourceChann

			go ql.AddListener(sourceChann)
		}
	}
}

func (ql *QueuesListener) AddListener(msgs <-chan amqp.Delivery) {
	for msg := range msgs {
		msgBody := bytes.NewReader(msg.Body)
		decoder := gob.NewDecoder(msgBody)
		sensorMsg := new(dto.SensorMessage)
		decoder.Decode(sensorMsg)

		fmt.Printf("Received sensor reading data message: %v\n", sensorMsg)

		// trigger an event
	}
}
