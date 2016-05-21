package coordinator

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/golang-distributed-application/src/powerplant/dto"
	"github.com/golang-distributed-application/src/powerplant/queueutils"
	"github.com/streadway/amqp"
)

const maxRate = 5 * time.Second

// !!! listening the events and determine which ones to forward to databse manager to persist the data
type DatabaseConsumer struct {
	er      EventRaiser // so DatabaseConsumer itself doesn't have to know how to publish event
	conn    *amqp.Connection
	ch      *amqp.Channel
	queue   *amqp.Queue // it's used to route the message to when DatabaseConsumer decides to persist in db
	sources []string
}

func NewDatabaseConsumer(er EventRaiser) *DatabaseConsumer {
	dc := DatabaseConsumer{
		er: er,
	}
	dc.conn, dc.ch = queueutils.GetChannel(url)
	dc.queue = queueutils.GetQueue(
		queueutils.PersistReadingsQueue, //name string,
		dc.ch, //ch *amqp.Channel,
		false) //autoDelete bool)

	dc.er.AddListener(queueutils.DataSourceDiscoveredEvent,
		// anonymous function
		func(eventData interface{}) {
			dc.SubscribeToDataEvent(eventData.(string))
		})

	return &dc
}

func (dc *DatabaseConsumer) SubscribeToDataEvent(eventName string) {
	for _, v := range dc.sources {
		if v == eventName {
			// existing data source
			return
		}
	}

	// !!! callback is a self-executing function that will return the callback itself
	// so a new isolated variable scope will be created every time i call this function
	// it's a trick to allow the event handler to register a bit of state that
	//  we'll need to throttle down the rate that the messages are coming in from the event sources.
	// here we throttle the messages to be persisted no faster than once every five seconds.
	dc.er.AddListener(queueutils.MessageReceivedEvent+eventName,
		func() func(interface{}) {
			prevTime := time.Unix(0, 0) // initial value, 45 years ago

			buf := new(bytes.Buffer)

			// !!! use a closure here since variables created in it are only visible to other members of the closure,
			// it means that prevTime and buf variables are going to be accessible by the returned callback function,
			// BUT they will retain their states from call to call since their states are captured by the closure, not the callback
			return func(eventData interface{}) {
				ed := eventData.(EventData)
				if time.Since(prevTime) > maxRate {
					prevTime = time.Now() // reset the prevTime

					// create a sensor message to publish
					sensorMsg := dto.SensorMessage{
						Name:      ed.Name,
						Value:     ed.Value,
						Timestamp: ed.Timestamp,
					}

					buf.Reset()

					encoder := gob.NewEncoder(buf)
					encoder.Encode(sensorMsg)

					// publishing
					msg := amqp.Publishing{
						Body: buf.Bytes(),
					}

					dc.ch.Publish(
						"", //exchange string,
						queueutils.PersistReadingsQueue, //key string,
						false, //mandatory bool,
						false, //immediate bool,
						msg)   //msg amqp.Publishing)
				}
			}
		}())
}
