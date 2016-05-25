package coordinator

import (
	"bytes"
	"encoding/gob"

	"github.com/golang-distributed-application/src/powerplant/dto"
	"github.com/golang-distributed-application/src/powerplant/queueutils"
	"github.com/streadway/amqp"
)

type WebappConsumer struct {
	er      EventRaiser
	conn    *amqp.Connection
	ch      *amqp.Channel
	sources []string
}

func NewWebappConsumer(er EventRaiser) *WebappConsumer {
	wc := WebappConsumer{
		er: er,
	}

	wc.conn, wc.ch = queueutils.GetChannel(url)
	queueutils.GetQueue(queueutils.PersistReadingsQueue, wc.ch, false)

	go wc.ListenForDiscoveryRequests()

	wc.er.AddListener(queueutils.DataSourceDiscoveredEvent,
		func(eventData interface{}) {
			wc.SubscribeToDataEvent(eventData.(string))
		})

	// declare exchanges
	/* !!!
	every coordinator will register to publish messages to these exchanges,
	and every web applications will register to listen to messages coming from them,
	we can easily create many-to-many relationship between coordinators and web applications,
	and horizontally scale.
	*/
	wc.ch.ExchangeDeclare(
		queueutils.WebappSourceExchange, //name string,
		"fanout",                        //kind string,
		false,                           //durable bool,
		false,                           //autoDelete bool,
		false,                           //internal bool,
		false,                           //noWait bool,
		nil)                             //args amqp.Table)

	wc.ch.ExchangeDeclare(
		queueutils.WebappReadingsExchange, //name string,
		"fanout", //kind string,
		false,    //durable bool,
		false,    //autoDelete bool,
		false,    //internal bool,
		false,    //noWait bool,
		nil)      //args amqp.Table)

	return &wc
}

func (wc *WebappConsumer) ListenForDiscoveryRequests() {
	queue := queueutils.GetQueue(queueutils.WebappDiscoveryQueue, wc.ch, false)
	msgs, _ := wc.ch.Consume(
		queue.Name, //queue string,
		"",         //consumer string,
		true,       //autoAck bool,
		false,      //exclusive bool,
		false,      //noLocal bool,
		false,      //noWait bool,
		nil)        //args amqp.Table)

	for range msgs {
		for _, src := range wc.sources {
			wc.SendMessageSource(src)
		}
	}
}

// SendMessageSource informs the web applications about the new sensor
func (wc *WebappConsumer) SendMessageSource(src string) {
	wc.ch.Publish(
		queueutils.WebappSourceExchange, //exchange string,
		"",    //key string,
		false, //mandatory bool,
		false, //immediate bool,
		amqp.Publishing{Body: []byte(src)}) // msg amqp.Publishing)
}

func (wc *WebappConsumer) SubscribeToDataEvent(eventName string) {
	for _, v := range wc.sources {
		if v == eventName {
			return
		}
	}

	wc.sources = append(wc.sources, eventName)

	wc.SendMessageSource(eventName)

	wc.er.AddListener(queueutils.MessageReceivedEvent+eventName,
		func(eventData interface{}) {
			ed := eventData.(EventData)
			sensorMsg := dto.SensorMessage{
				Name:      ed.Name,
				Value:     ed.Value,
				Timestamp: ed.Timestamp,
			}

			buffer := new(bytes.Buffer)
			encoder := gob.NewEncoder(buffer)
			encoder.Encode(sensorMsg)
			msg := amqp.Publishing{
				Body: buffer.Bytes(),
			}

			wc.ch.Publish(
				queueutils.WebappReadingsExchange, //exchange string,
				"",    //key string,
				false, //mandatory bool,
				false, //immediate bool,
				msg)   //msg amqp.Publishing)
		})
}
