package context

import (
	"fmt"
	"strings"
	"time"
)

// DebugNode ...
type DebugNode struct {
	ID            int64
	ComponentType string
	ComponentName string
}

// Debugger ...
type Debugger interface {
	Log(nodePath []DebugNode, vars ...interface{})
}

// EmptyDebugger ...
type EmptyDebugger struct{}

// Log ...
func (emptyDebugger *EmptyDebugger) Log(nodePath []DebugNode, vars ...interface{}) {}

// NewEmptyDebugger ...
func NewEmptyDebugger() Debugger {
	return &EmptyDebugger{}
}

// ConsoleLogDebugger ...
type ConsoleLogDebugger struct{}

// NewConsoleLogDebugger ...
func NewConsoleLogDebugger() Debugger {
	return &ConsoleLogDebugger{}
}

// Log ...
func (consoleLogDebugger *ConsoleLogDebugger) Log(nodePath []DebugNode, arguments ...interface{}) {
	pathStrings := []string{}

	for _, node := range nodePath {
		pathStrings = append(pathStrings, node.ComponentName)
	}

	path := strings.Join(pathStrings, "->")

	values := []string{}
	values = append(values, fmt.Sprintf("%v", time.Now().Format(time.RFC3339)))
	values = append(values, path)

	valuesStr := strings.Join(values, ",")

	vars := []string{}
	for _, argument := range arguments {
		for _, parameter := range argument.([]interface{}) {
			vars = append(vars, fmt.Sprintf("%v", parameter))
		}
	}
	varsStr := strings.Join(vars, ",")

	fmt.Println(fmt.Sprintf("%v    %v", valuesStr, varsStr))
}
