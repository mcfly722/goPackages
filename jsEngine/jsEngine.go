package jsEngine

import (
	"fmt"
	"log"
	"sync"

	"github.com/dop251/goja"
)

// JSRuntime ...
type JSRuntime struct {
	Name      string
	EventLoop chan *event
	VM        *goja.Runtime
}

func (runtime *JSRuntime) startEventLoop(jsContent string) error {
	_, err := runtime.VM.RunString(jsContent)
	if err != nil {
		return err
	}

	go func() {
		log.Printf(fmt.Sprintf("%v started", runtime.Name))

	loop:
		for {
			select {
			case event := <-runtime.EventLoop:
				if event.kind == 0 {
					break loop
				}
			}
		}

		log.Printf(fmt.Sprintf("%v stopped", runtime.Name))
	}()

	return nil
}

func (runtime *JSRuntime) close() {
	runtime.EventLoop <- &event{
		kind: 0,
	}
}

// JSEngine ...
type JSEngine struct {
	runtimes     map[string]*JSRuntime
	apiFunctions [](func(runtime *JSRuntime) error)
	ready        sync.Mutex
}

func (jsEngine *JSEngine) newRuntime(name string, jsContent string) (*JSRuntime, error) {

	runtime := &JSRuntime{
		Name:      name,
		EventLoop: make(chan *event),
		VM:        goja.New(),
	}

	// apply all api functions to vm
	for _, apiFunction := range jsEngine.apiFunctions {
		if err := apiFunction(runtime); err != nil {
			return nil, err
		}
	}

	return runtime, runtime.startEventLoop(jsContent)
}

// event from JSEngine to Goroutine with JS Loop
type event struct {
	kind int
	// 0 -exit (finish go-routine)
	data string
}

// NewJSEngine ...
func NewJSEngine() *JSEngine {
	return &JSEngine{
		runtimes:     make(map[string](*JSRuntime)),
		apiFunctions: [](func(*JSRuntime) error){},
	}
}

// NewRuntime ...
func (jsEngine *JSEngine) NewRuntime(name string, jsContent string) error {
	jsEngine.ready.Lock()
	_, found := jsEngine.runtimes[name]

	if found {
		jsEngine.ready.Unlock()
		return fmt.Errorf("namespace=%v already exist", name)
	}

	// add new Namespace
	runtime, err := jsEngine.newRuntime(name, jsContent)
	if err != nil {
		jsEngine.ready.Unlock()
		return fmt.Errorf("namespace=%v error:%v", name, err)
	}

	jsEngine.runtimes[name] = runtime

	jsEngine.ready.Unlock()

	return nil
}

// CloseRuntime ...
func (jsEngine *JSEngine) CloseRuntime(name string) error {
	jsEngine.ready.Lock()

	if runtime, found := jsEngine.runtimes[name]; found {

		runtime.close()

		delete(jsEngine.runtimes, name)

		jsEngine.ready.Unlock()
	} else {
		jsEngine.ready.Unlock()
		return fmt.Errorf("namespace=%v not found for finishing", name)
	}
	return nil
}

// AddAPI ...
func (jsEngine *JSEngine) AddAPI(apiFunction func(*JSRuntime) error) {
	jsEngine.apiFunctions = append(jsEngine.apiFunctions, apiFunction)
}
