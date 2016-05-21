package coordinator

import "time"

// EventRaiser will be used in the consumer, so the consumer itself doesn't have to know how to publish the event
type EventRaiser interface {
	AddListener(eventName string, f func(interface{})) // change the parameter to generic type, too.
}

// EventAggregator lets consumers use AddListener to register an event, then it'll loop through all the events and
// trigger the callback function for each registered consumer
type EventAggregator struct {
	listeners map[string][]func(interface{}) // map's value is all the callbacks for all registered consumers for this specific event
}

type EventData struct {
	Name      string
	Value     float64
	Timestamp time.Time
}

func NewEventAggregator() *EventAggregator {
	ea := EventAggregator{
		listeners: make(map[string][]func(interface{})),
	}

	return &ea
}

// AddListener lets a consumer register event(name) with its callback function(callback)
func (ea *EventAggregator) AddListener(name string, callback func(interface{})) {
	ea.listeners[name] = append(ea.listeners[name], callback)
}

// PublishEvent loops all the registered consumers' callbacks and trigger the same event for each of them
// NOTE: eventData is passed by value since it should be mutable.
func (ea *EventAggregator) PublishEvent(name string, eventData interface{}) {
	if ea.listeners[name] != nil {
		for _, callback := range ea.listeners[name] {
			callback(eventData)
		}
	}
}
