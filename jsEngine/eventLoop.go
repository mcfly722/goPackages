package jsEngine

import (
	"fmt"

	"github.com/dop251/goja"
	"github.com/mcfly722/goPackages/context"
)

// EventLoop ...
type EventLoop interface {
	context.ContextedInstance
	AddAPI(api APIConstructor)
	CallHandler(function *goja.Callable, args ...goja.Value) (goja.Value, error)
}

// Script ...
type Script interface {
	getBody() string
	getName() string
}

// APIConstructor ...
type APIConstructor func(context context.Context, eventLoop EventLoop, runtime *goja.Runtime)

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
	apis     []APIConstructor
	scripts  []Script
	handlers chan *handler
}

func (eventLoop *eventLoop) AddAPI(api APIConstructor) {
	eventLoop.apis = append(eventLoop.apis, api)
}

// NewEventLoop ...
func NewEventLoop(runtime *goja.Runtime, scripts []Script) EventLoop {
	eventLoop := &eventLoop{
		runtime:  runtime,
		apis:     []APIConstructor{},
		scripts:  scripts,
		handlers: make(chan *handler),
	}

	return eventLoop
}

// Go ...
func (eventLoop *eventLoop) Go(current context.Context) {

	for _, api := range eventLoop.apis {
		api(current, eventLoop, eventLoop.runtime)
	}

	for _, script := range eventLoop.scripts {
		_, err := eventLoop.runtime.RunString(script.getBody())
		if err != nil {
			current.Log(1, fmt.Sprintf("%v: %v", script.getName(), err.Error()))
			return
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
