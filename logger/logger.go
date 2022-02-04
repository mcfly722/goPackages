package logger

import (
	"fmt"
	"sync"
	"time"
)

var (
	// EventTypes ...
	EventTypes = map[int]string{0: "EXPN", 1: "TRCE", 2: "INFO"}
)

const (
	// EventTypeException ...
	EventTypeException = 0
	// EventTypeTrace ...
	EventTypeTrace = 1
	// EventTypeInfo ...
	EventTypeInfo = 0
)

// Event ...
type Event struct {
	DateTime time.Time
	Number   int64
	Type     int
	Object   string
	Message  string
}

// Logger ...
type Logger struct {
	events  []*Event
	counter int64

	ready sync.Mutex
}

// NewLogger ...
func NewLogger(bufferSize int) *Logger {
	return &Logger{
		events:  make([]*Event, bufferSize),
		counter: 0,
	}
}

// LogEvent ...
func (logger *Logger) LogEvent(eventType int, object string, message string) {

	logger.ready.Lock()

	newEvent := &Event{
		DateTime: time.Now(),
		Number:   logger.counter,
		Type:     eventType,
		Object:   object,
		Message:  message,
	}

	eventIndex := (int)(logger.counter % (int64)(len(logger.events)))
	logger.events[eventIndex] = newEvent

	logger.counter++

	logger.ready.Unlock()
}

// EventTypeToText ...
func EventTypeToText(eventType int) string {
	if value, ok := EventTypes[eventType]; ok {
		return value
	}
	return fmt.Sprintf("%4v", eventType)
}

// EventsTextRepresentation ...
func EventsTextRepresentation(events *[]Event) string {
	result := ""
	for _, event := range *events {
		result = result + fmt.Sprintf("%v %5v [%v] %v: %v\n", event.DateTime.Format(time.RFC3339), event.Number, EventTypeToText(event.Type), event.Object, event.Message)
	}
	return result
}

// GetLastEvents ...
func (logger *Logger) GetLastEvents(startFrom int64) *[]Event {

	result := []Event{}

	logger.ready.Lock()
	if logger.counter > 0 {

		for i := logger.counter - 1; i >= startFrom && i >= logger.counter-(int64)(len(logger.events)) && i > -1; i-- {
			eventIndex := (int)(i % (int64)(len(logger.events)))

			p := logger.events[eventIndex]
			if p != nil {
				result = append(result, *p)
			}
		}
	}
	logger.ready.Unlock()

	resultSorted := []Event{}

	for i := len(result) - 1; i >= 0; i-- {
		resultSorted = append(resultSorted, result[i])
	}

	return &resultSorted
}
