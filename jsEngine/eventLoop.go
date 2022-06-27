package jsEngine

import (
	"fmt"

	"github.com/dop251/goja"
	"github.com/mcfly722/goPackages/context"
)

type apiConstructor func(context context.Context, eventLoop *eventLoop, runtime *goja.Runtime)

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

func newScript(name string, body string) *script {
	return &script{
		name: name,
		body: body,
	}
}

type eventLoop struct {
	runtime  *goja.Runtime
	apis     []apiConstructor
	scripts  []*script
	handlers chan *handler
}

func (eventLoop *eventLoop) addAPI(api apiConstructor) {
	eventLoop.apis = append(eventLoop.apis, api)
}

func newEventLoop(runtime *goja.Runtime, scripts []*script) *eventLoop {
	eventLoop := &eventLoop{
		runtime:  runtime,
		apis:     []apiConstructor{},
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
		_, err := eventLoop.runtime.RunString(script.body)
		if err != nil {
			current.Log(1, fmt.Sprintf("%v: %v", script.name, err.Error()))
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

func (eventLoop *eventLoop) Call(function *goja.Callable, args ...goja.Value) (goja.Value, error) {

	results := make(chan result)

	eventLoop.handlers <- &handler{
		function:      function,
		args:          args,
		resultChannel: results,
	}

	result := <-results

	return result.value, result.err
}
