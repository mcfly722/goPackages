package jsEngine

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/dop251/goja"
)

// JSEngine ...
type JSEngine struct {
	runtimesChannels map[string](chan *event)
	api              func(vm *goja.Runtime)
	ready            sync.Mutex
}

func (jsEngine *JSEngine) newRuntime(name string, jsContent string) (chan *event, error) {
	channel := make(chan *event)

	vm := goja.New()

	jsEngine.api(vm)

	_, err := vm.RunString(jsContent)
	if err != nil {
		return nil, err
	}

	go func() {
		log.Printf(fmt.Sprintf("%v started", name))

	loop:
		for {
			select {
			case event := <-channel:
				if event.kind == 0 {
					break loop
				}

			default:
				// loop
				time.Sleep(100 * time.Millisecond)
				fmt.Printf(".")
			}
		}

		log.Printf(fmt.Sprintf("%v stopped", name))
	}()

	return channel, nil
}

// event from JSEngine to Goroutine with JS Loop
type event struct {
	kind int
	// 0 -exit (finish go-routine)
	data string
}

// NewJSEngine ...
func NewJSEngine(api func(vm *goja.Runtime)) *JSEngine {
	return &JSEngine{
		runtimesChannels: make(map[string](chan *event)),
		api:              api,
	}
}

// NewRuntime ...
func (jsEngine *JSEngine) NewRuntime(name string, jsContent string) error {
	jsEngine.ready.Lock()
	_, found := jsEngine.runtimesChannels[name]

	if found {
		jsEngine.ready.Unlock()
		return fmt.Errorf("namespace=%v already exist", name)
	}

	// add new Namespace
	ch, err := jsEngine.newRuntime(name, jsContent)
	if err != nil {
		jsEngine.ready.Unlock()
		return fmt.Errorf("namespace=%v error:%v", name, err)
	}

	jsEngine.runtimesChannels[name] = ch

	jsEngine.ready.Unlock()

	return nil
}

// CloseRuntime ...
func (jsEngine *JSEngine) CloseRuntime(name string) error {
	jsEngine.ready.Lock()

	if ch, found := jsEngine.runtimesChannels[name]; found {

		ch <- &event{
			kind: 0,
		}

		delete(jsEngine.runtimesChannels, name)

		jsEngine.ready.Unlock()
	} else {
		jsEngine.ready.Unlock()
		return fmt.Errorf("namespace=%v not found for finishing", name)
	}
	return nil
}
