package logger

import (
	"fmt"
	"math/rand"
	"testing"
)

func Test_LogUnknownEventType(t *testing.T) {
	logger := NewLogger(10)

	logger.LogEvent(660, "OBJECT#0", "Message#0")
	logger.LogEvent(661, "OBJECT#1", "Message#1")
	logger.LogEvent(662, "OBJECT#2", "Message#2")
	logger.LogEvent(663, "OBJECT#3", "Message#3")

	fmt.Printf(EventsTextRepresentation(logger.GetLastEvents(0)))

}

func Test_LogOverflowEventType(t *testing.T) {
	logger := NewLogger(3)

	logger.LogEvent(EventTypeInfo, "OBJECT#0", "Message#0")
	logger.LogEvent(EventTypeInfo, "OBJECT#1", "Message#1")
	logger.LogEvent(EventTypeInfo, "OBJECT#2", "Message#2")
	logger.LogEvent(EventTypeInfo, "OBJECT#3", "Message#3")
	logger.LogEvent(EventTypeInfo, "OBJECT#4", "Message#4")
	logger.LogEvent(EventTypeInfo, "OBJECT#5", "Message#5")

	fmt.Printf(EventsTextRepresentation(logger.GetLastEvents(0)))
}

func Test_LogToManyNewEvents(t *testing.T) {
	logger := NewLogger(4)

	logger.LogEvent(EventTypeInfo, "OBJECT#0", "Message#0")
	logger.LogEvent(EventTypeInfo, "OBJECT#1", "Message#1")
	logger.LogEvent(EventTypeInfo, "OBJECT#2", "Message#2")
	logger.LogEvent(EventTypeInfo, "OBJECT#3", "Message#3")
	logger.LogEvent(EventTypeInfo, "OBJECT#4", "Message#4")
	logger.LogEvent(EventTypeInfo, "OBJECT#5", "Message#5")

	fmt.Printf(EventsTextRepresentation(logger.GetLastEvents(0)))
}

func Test_Log(t *testing.T) {

	logger := NewLogger(10)

	for thread := 0; thread < 7; thread++ {
		threadNumber := thread
		go func() {
			for i := 0; i < 1000; i++ {
				logger.LogEvent(rand.Intn(3), fmt.Sprintf("object#%v", threadNumber), fmt.Sprintf("message#%v", i))
			}

			fmt.Printf(fmt.Sprintf("thread %2v -----------\n%v", threadNumber, EventsTextRepresentation(logger.GetLastEvents(0))))
		}()
	}

}

func Test_ZeroLog(t *testing.T) {
	logger := NewLogger(0)
	logger.LogEvent(EventTypeInfo, "OBJECT#0", "Message#0")
	logger.LogEvent(EventTypeInfo, "OBJECT#1", "Message#1")
	logger.LogEvent(EventTypeInfo, "OBJECT#2", "Message#2")
	fmt.Printf(EventsTextRepresentation(logger.GetLastEvents(0)))
}

func Test_ZeroLogWithoutOutput(t *testing.T) {
	logger := NewLogger(0)
	logger.SetOutputToConsole(false)
	if logger.IsOutputToConsoleEnabled() {
		t.Fatal("IsOutputToConsoleEnabled not works!")
	}
	logger.LogEvent(EventTypeInfo, "OBJECT#0", "Message#0")
	logger.LogEvent(EventTypeInfo, "OBJECT#1", "Message#1")
	logger.LogEvent(EventTypeInfo, "OBJECT#2", "Message#2")
	fmt.Printf(EventsTextRepresentation(logger.GetLastEvents(0)))
}
