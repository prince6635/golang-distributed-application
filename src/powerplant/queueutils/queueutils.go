package queueutils

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

// SensorListQueue is the queue to record all sensor queues' names
const SensorListQueue = "SensorList"

// SensorDiscoveryExchange is the exchange name for queue listeners to discover previous started data queues if the coordinator started after the data queues
const SensorDiscoveryExchange = "SensorDiscovery"
const PersistReadingsQueue = "PersistReadings"

// WebappSourceExchange is used to broadcast the name of the sensors
const WebappSourceExchange = "WebappSources"

// WebappReadingsExchange is used to send out readings to the web applications.
const WebappReadingsExchange = "WebappReadings"

// WebappDiscoveryQueue is used by the web applications to let the coordinator know
// 	that one of them would like to get a list of all of the available sources.
const WebappDiscoveryQueue = "WebappDiscovery"

// event names
const DataSourceDiscoveredEvent = "DataSourceDiscovered"
const MessageReceivedEvent = "MessageReceived_"

// GetChannel returns the connection and channel from RabbitMQ
func GetChannel(url string) (*amqp.Connection, *amqp.Channel) {
	conn, err := amqp.Dial(url)
	failOnError(err, "Failed to establish connection to message broker.")

	ch, err := conn.Channel()
	failOnError(err, "Failed to get channel for connection")

	return conn, ch
}

// GetQueue returns a queue from RabbitMQ
func GetQueue(name string, ch *amqp.Channel, autoDelete bool) *amqp.Queue {
	q, err := ch.QueueDeclare(
		name,  //name string,
		false, //durable bool,
		// !!! true will clean up the temp queues
		autoDelete, //autoDelete bool,
		false,      //exclusive bool,
		false,      //noWait bool,
		nil)        //args amqp.Table)

	failOnError(err, "Failed to declare a queue")
	return &q
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}
