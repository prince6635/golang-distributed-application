package controller

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net/http"
	"sync"

	"github.com/golang-distributed-application/src/powerplant/dto"
	"github.com/golang-distributed-application/src/powerplant/queueutils"
	"github.com/golang-distributed-application/src/powerplant/web/model"
	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
)

const url = "amqp://guest:guest@localhost:5672"

type websocketController struct {
	conn     *amqp.Connection
	ch       *amqp.Channel
	sockets  []*websocket.Conn
	mutex    sync.Mutex
	upgrader websocket.Upgrader // upgrade specially formed http request to a web socket
}

func newWebsocketController() *websocketController {
	wsc := new(websocketController)

	wsc.conn, wsc.ch = queueutils.GetChannel(url)

	wsc.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	// send two types of messages we're getting from RabbitMQ to the web clients
	go wsc.listenForSources()
	go wsc.listenForMessages()

	return wsc
}

// handler function that actually handles http request that being received by the controller
func (wsc *websocketController) handleMessage(w http.ResponseWriter, r *http.Request) {
	socket, _ := wsc.upgrader.Upgrade(w, r, nil)
	wsc.addSocket(socket)
	go wsc.listenForDiscoveryRequests(socket)
}

func (wsc *websocketController) addSocket(socket *websocket.Conn) {
	wsc.mutex.Lock()
	wsc.sockets = append(wsc.sockets, socket)
	wsc.mutex.Unlock()
}

func (wsc *websocketController) removeSocket(socket *websocket.Conn) {
	wsc.mutex.Lock()
	socket.Close()

	for i := range wsc.sockets {
		if wsc.sockets[i] == socket {
			wsc.sockets = append(wsc.sockets[:i], wsc.sockets[i+1:]...)
		}
	}

	wsc.mutex.Unlock()
}

func (wsc *websocketController) listenForDiscoveryRequests(socket *websocket.Conn) {
	for {
		msg := message{}
		err := socket.ReadJSON(&msg)

		if err != nil {
			// it means the socket might be corrupted or closed from the client's end
			wsc.removeSocket(socket)
			break
		}

		if msg.Type == "discover" {
			// !!! this will be picked up by one of coordinators and respond with a list of sensors
			wsc.ch.Publish(
				"", //exchange string,
				queueutils.WebappDiscoveryQueue, //key string,
				false,             //mandatory bool,
				false,             //immediate bool,
				amqp.Publishing{}) //msg amqp.Publishing)
		}
	}
}

// send the message to ALL the web clients
func (wsc *websocketController) sendMessage(msg message) {
	// !!! remove the dead sockets if sending failed
	socketsToRemove := []*websocket.Conn{}

	for _, socket := range wsc.sockets {
		err := socket.WriteJSON(msg)

		if err != nil {
			socketsToRemove = append(socketsToRemove, socket)
		}
	}

	for _, socket := range socketsToRemove {
		wsc.removeSocket(socket)
	}
}

type message struct {
	// tell the Data's type, since it's always empty interface
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func (wsc *websocketController) listenForSources() {
	q := queueutils.GetQueue("", wsc.ch, true)
	wsc.ch.QueueBind(
		q.Name, //name string,
		"",     //key string,
		queueutils.WebappSourceExchange, //exchange string,
		false, //noWait bool,
		nil)   //args amqp.Table)

	msgs, _ := wsc.ch.Consume(
		q.Name, //queue string,
		"",     //consumer string,
		true,   //autoAck bool,
		false,  //exclusive bool,
		false,  //noLocal bool,
		false,  //noWait bool,
		nil)    //args amqp.Table)

	for msg := range msgs {
		sensor, err := model.GetSensorByName(string(msg.Body))
		if err != nil {
			fmt.Println(err.Error())
		}
		wsc.sendMessage(message{
			Type: "source",
			Data: sensor,
		})
	}
}

func (wsc *websocketController) listenForMessages() {
	q := queueutils.GetQueue("", wsc.ch, true)
	wsc.ch.QueueBind(
		q.Name, //name string,
		"",     //key string,
		queueutils.WebappReadingsExchange, //exchange string,
		false, //noWait bool,
		nil)   //args amqp.Table)

	msgs, _ := wsc.ch.Consume(
		q.Name, //queue string,
		"",     //consumer string,
		true,   //autoAck bool,
		false,  //exclusive bool,
		false,  //noLocal bool,
		false,  //noWait bool,
		nil)    //args amqp.Table)

	for msg := range msgs {
		buffer := bytes.NewBuffer(msg.Body)
		decoder := gob.NewDecoder(buffer)
		sensorMsg := dto.SensorMessage{}
		err := decoder.Decode(&sensorMsg)

		if err != nil {
			fmt.Println(err.Error())
		}

		wsc.sendMessage(message{
			Type: "reading",
			Data: sensorMsg,
		})
	}
}
