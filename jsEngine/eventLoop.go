package jsEngine

import (
	"fmt"

	"github.com/dop251/goja"
	"github.com/mcfly722/goPackages/context"
)

// EventLoop ...
type EventLoop interface {
	context.ContextedInstance
	Import(module Module)
	CallHandler(function *goja.Callable, args ...goja.Value) (goja.Value, error)
}

// Script ...
type Script interface {
	getBody() string
	getName() string
}

// Module ...
type Module interface {
	Constructor(context context.Context, eventLoop EventLoop, runtime *goja.Runtime)
}

type result struct {
	value goja.Value
	err   error
}

type handler struct {
	function      *goja.Callable
	args          []goja.Value
	resultChannel chan result
}

type script struct {
	name string
	body string
}

func (script *script) getBody() string {
	return script.body
}

func (script *script) getName() string {
	return script.name
}

// NewScript ...
func NewScript(name string, body string) Script {
	return &script{
		name: name,
		body: body,
	}
}

type eventLoop struct {
	runtime  *goja.Runtime
	modules  []Module
	scripts  []Script
	handlers chan *handler
}

// Import ...
func (eventLoop *eventLoop) Import(module Module) {
	eventLoop.modules = append(eventLoop.modules, module)
}

// NewEventLoop ...
func NewEventLoop(runtime *goja.Runtime, scripts []Script) EventLoop {
	eventLoop := &eventLoop{
		runtime:  runtime,
		modules:  []Module{},
		scripts:  scripts,
		handlers: make(chan *handler),
	}

	return eventLoop
}

// Go ...
func (eventLoop *eventLoop) Go(current context.Context) {

	for _, module := range eventLoop.modules {
		module.Constructor(current, eventLoop, eventLoop.runtime)
	}

	for _, script := range eventLoop.scripts {
		_, err := eventLoop.runtime.RunString(script.getBody())
		if err != nil {
			current.Log(1, fmt.Sprintf("%v: %v", script.getName(), err.Error()))
			current.Cancel()
		}
	}

loop:
	for {
		select {

		case handler := <-eventLoop.handlers:
			value, err := (*handler.function)(nil, handler.args...)

			handler.resultChannel <- result{
				value: value,
				err:   err,
			}

			break
		case _, opened := <-current.Opened():
			if !opened {
				break loop
			}
		}
	}
}

func (eventLoop *eventLoop) CallHandler(function *goja.Callable, args ...goja.Value) (goja.Value, error) {

	results := make(chan result)

	eventLoop.handlers <- &handler{
		function:      function,
		args:          args,
		resultChannel: results,
	}

	result := <-results

	return result.value, result.err
}
