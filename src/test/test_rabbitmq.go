package test

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

// RabbitMQ tests the publisher and consumer of RabbitMQ
func RabbitMQ() {
	// test RabbitMQ
	go publish()
	go consume()

	// keep goroutines running
	var hang string
	fmt.Scanln(&hang)
}

// Publishes message to RabbitMQ
func publish() {
	conn, ch, q := getQueue()
	defer conn.Close()
	defer ch.Close()

	msg := amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte("Hello RabbitMQ"),
	}

	for {
		ch.Publish("", q.Name, false, false, msg)
	}
}

// Consumes message from RabbitMQ
func consume() {
	conn, ch, q := getQueue()
	defer conn.Close()
	defer ch.Close()

	msgs, err := ch.Consume(
		q.Name, //queue string,
		"",     // consumer string,
		true,   //autoAck bool,
		false,  //exclusive bool,
		false,  //noLocal bool,
		false,  //noWait bool,
		nil)    //args amqp.Table)

	failOnError(err, "Failed to register a consumer")

	for msg := range msgs {
		log.Printf("Received message: %s\n", msg.Body)
	}
}

func getQueue() (*amqp.Connection, *amqp.Channel, *amqp.Queue) {
	conn, err := amqp.Dial("amqp://guest@localhost:5672")
	failOnError(err, "Failed to connect to RabbitMQ")
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	q, err := ch.QueueDeclare("test_queue",
		false, //durable bool,
		false, //autoDelete bool,
		false, //exclusive bool,
		false, //noWait bool,
		nil)   //args amqp.Table)
	failOnError(err, "Failed to declare a queue")

	return conn, ch, &q
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}
