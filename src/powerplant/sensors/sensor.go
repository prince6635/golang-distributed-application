package sensor

import (
	"bytes"
	"encoding/gob"
	"flag"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/golang-distributed-application/src/powerplant/dto"
	"github.com/golang-distributed-application/src/powerplant/queueutils"
	"github.com/streadway/amqp"
)

// url for RabbitMQ channel, should be inside a config file
var url = "amqp://guest:guest@localhost:5672"

/* flag:
Command-line flags are a common way to specify options for command-line programs.
For example, in wc -l the -l is a command-line flag.
Go provides a flag package supporting basic command-line flag parsing.
*/
// fields related to calculate the sensor's value
var name = flag.String("name", "sensor", "name of the sensor")
var frequency = flag.Uint("freq", 5, "update frequency in cycles/sec")
var max = flag.Float64("max", 5., "maximum value for generated readings")
var min = flag.Float64("min", 1., "minimum value for generated readings")
var stepSize = flag.Float64("step", 0.1, "maximum allowable change per measurement")

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

var value = random.Float64()*(*max-*min) + *min
var normalValue = (*max-*min)/2 + *min

// calculate the sensor's next value
func getNextSensorValue() {
	var maxStep, minStep float64

	if value < normalValue {
		maxStep = *stepSize
		minStep = -1 * *stepSize * (value - *min) / (normalValue - *min)
	} else {
		maxStep = *stepSize * (*max - value) / (*max - normalValue)
		minStep = -1 * *stepSize
	}

	value += random.Float64()*(maxStep-minStep) + minStep
}

// StartPublishingSensorData publishes data from sensors to RabbitMQ
func StartPublishingSensorData() {
	flag.Parse()

	conn, ch := queueutils.GetChannel(url)
	defer conn.Close()
	defer ch.Close()

	publishSensorNameToSensorListQueue(ch, *name)

	publishSensorDataToSensorQueue(ch)
}

// record each sensor queue's name into sensor list queue
/* Sample data:
Exchange: (AMQP default)
Routing Key: SensorList
Payload:
	6 bytes
	Encoding: string
	sensor
*/
func publishSensorNameToSensorListQueue(ch *amqp.Channel, sensorName string) {
	sensorListQueue := queueutils.GetQueue(queueutils.SensorListQueue, ch)
	msg := amqp.Publishing{Body: []byte(sensorName)}
	ch.Publish(
		"",                   //exchange string,
		sensorListQueue.Name, //key string,
		false,                //mandatory bool,
		false,                //immediate bool,
		msg)                  //msg amqp.Publishing)
}

// record sensor reading data into sensor queue
/* sampel data:
Exchange:	(AMQP default)
Routing Key:	sensor
Payload:
	118 bytes
	Encoding: base64
	Pf+BAwEBDVNlbnNvck1lc3NhZ2UB/4IAAQMBBE5hbWUBDAABBVZhbHVlAQgAAQlUaW1lc3RhbXAB/4QAAAAQ/4MFAQEEVGltZQH/hAAAACb/ggEGc2Vuc29y
	Afhp5QFYPncQQAEPAQAAAA7OuQuXF2Cyw/5cAA==
*/
func publishSensorDataToSensorQueue(ch *amqp.Channel) {
	sensorDataQueue := queueutils.GetQueue(*name, ch)

	duration, _ := time.ParseDuration(strconv.Itoa(1000/int(*frequency)) + "ms")
	signal := time.Tick(duration)
	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)

	// publish sensor messages
	for range signal {
		getNextSensorValue()

		sensorReadingMsg := dto.SensorMessage{
			Name:      *name,
			Value:     value,
			Timestamp: time.Now(),
		}

		buffer.Reset()
		encoder.Encode(sensorReadingMsg)

		msg := amqp.Publishing{
			Body: buffer.Bytes(),
		}

		ch.Publish(
			"",                   //exchange string,
			sensorDataQueue.Name, //key string,
			false,                //mandatory bool,
			false,                //immediate bool,
			msg)                  //msg amqp.Publishing)

		log.Printf("Sensor reading message sent, value: %v\n", value)
	}
}
