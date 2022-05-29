package logger

import (
	"fmt"
	"sync"
	"time"
)

const (
	// EventType ...
	EventType = 0

	// EventTypeException ...
	EventTypeException = 1
	// EventTypeInfo ...
	EventTypeInfo = 2
	// EventTypeTrace ...
	EventTypeTrace = 3
)

var (
	// EventTypes ...
	EventTypes = map[int]string{
		EventType:          "    ",
		EventTypeException: "EXPN",
		EventTypeInfo:      "INFO",
		EventTypeTrace:     "TRCE",
	}
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
	events        []*Event
	counter       int64
	consoleOutput bool
	ready         sync.Mutex
}

// NewLogger ...
func NewLogger(bufferSize int) *Logger {
	return &Logger{
		events:        make([]*Event, bufferSize),
		consoleOutput: true,
		counter:       0,
	}
}

// SetOutputToConsole ...
func (logger *Logger) SetOutputToConsole(flag bool) {
	logger.ready.Lock()
	logger.consoleOutput = flag
	logger.ready.Unlock()
}

// IsOutputToConsoleEnabled ...
func (logger *Logger) IsOutputToConsoleEnabled() bool {
	logger.ready.Lock()
	result := logger.consoleOutput
	logger.ready.Unlock()
	return result
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

	if len(logger.events) > 0 {
		eventIndex := (int)(logger.counter % (int64)(len(logger.events)))
		logger.events[eventIndex] = newEvent
	}

	if logger.consoleOutput {
		fmt.Println(newEvent.ToString())
	}

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

// ToString ...
func (event *Event) ToString() string {
	return fmt.Sprintf("%v %8v [%v] %v: %v", event.DateTime.Format(time.RFC3339), event.Number, EventTypeToText(event.Type), event.Object, event.Message)
}

// EventsTextRepresentation ...
func EventsTextRepresentation(events *[]Event) string {
	result := ""
	for _, event := range *events {
		result = result + event.ToString() + "\n"
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
