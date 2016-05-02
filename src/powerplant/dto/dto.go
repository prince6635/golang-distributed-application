package dto

import (
	"encoding/gob"
	"time"
)

// SensorMessage represents the message that's sent from a sensor
type SensorMessage struct {
	Name      string
	Value     float64
	Timestamp time.Time
}

func int() {
	gob.Register(SensorMessage{})
}
