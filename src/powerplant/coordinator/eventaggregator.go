package coordinator

import "time"

// EventAggregator lets consumers use AddListener to register an event, then it'll loop through all the events and
// trigger the callback function for each registered consumer
type EventAggregator struct {
	listeners map[string][]func(EventData) // map's value is all the callbacks for all registered consumers for this specific event
}

type EventData struct {
	Name      string
	Value     float64
	Timestamp time.Time
}

func NewEventAggregator() *EventAggregator {
	ea := EventAggregator{
		listeners: make(map[string][]func(EventData)),
	}

	return &ea
}

// AddListener lets a consumer register event(name) with its callback function(callback)
func (ea *EventAggregator) AddListener(name string, callback func(EventData)) {
	ea.listeners[name] = append(ea.listeners[name], callback)
}

// PublishEvent loops all the registered consumers' callbacks and trigger the same event for each of them
// NOTE: eventData is passed by value since it should be mutable.
func (ea *EventAggregator) PublishEvent(name string, eventData EventData) {
	if ea.listeners[name] != nil {
		for _, callback := range ea.listeners[name] {
			callback(eventData)
		}
	}
}
