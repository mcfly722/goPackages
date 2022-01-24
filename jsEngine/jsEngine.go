package jsEngine

import (
	"fmt"
	"log"
	"sync"

	"github.com/dop251/goja"
)

// JSRuntime ...
type JSRuntime struct {
	Name           string
	EventLoop      chan *event
	apiFunctions   [](APIFunction)
	VM             *goja.Runtime
	jsContent      string
	apiInitialized bool
	destroyed      bool

	ready sync.Mutex
}

// APIFunction ...
type APIFunction interface {
	Initialize(jsRuntime *JSRuntime) error
	Dispose(jsRuntime *JSRuntime)
}

func (runtime *JSRuntime) initAPI() error {
	startedAPIFunctions := [](APIFunction){}

	// apply all api functions to vm, or rollback all api using dispose
	for _, apiFunction := range runtime.apiFunctions {
		if err := apiFunction.Initialize(runtime); err != nil {

			for _, initializedFunction := range startedAPIFunctions {
				initializedFunction.Dispose(runtime)
			}

			return err
		}

		startedAPIFunctions = append(startedAPIFunctions, apiFunction)
	}

	return nil
}

// Start ...
func (runtime *JSRuntime) Start() error {

	if runtime.destroyed {
		return fmt.Errorf("%v runtime already destroyed", runtime.Name)
	}

	if !runtime.apiInitialized {
		if err := runtime.initAPI(); err != nil {
			return err
		}
		runtime.apiInitialized = true
	}

	log.Printf(fmt.Sprintf("%v started", runtime.Name))
	// start content execution
	_, err := runtime.VM.RunString(runtime.jsContent)
	if err != nil {
		return err
	}

	go func() {

	loop:
		for {
			select {
			case event := <-runtime.EventLoop:
				if event.kind == 0 {
					break loop
				}
			}
		}

		log.Printf(fmt.Sprintf("%v finished", runtime.Name))
	}()

	return nil
}

// Stop ...
func (runtime *JSRuntime) destroy() {

	if runtime.apiInitialized {
		for _, apiFunction := range runtime.apiFunctions {
			apiFunction.Dispose(runtime)
		}
	}

	if !runtime.destroyed && runtime.apiInitialized {
		runtime.EventLoop <- &event{
			kind: 0,
		}
	}

	runtime.destroyed = true
}

// JSEngine ...
type JSEngine struct {
	runtimes map[string]*JSRuntime
	ready    sync.Mutex
}

func (jsEngine *JSEngine) newRuntime(name string, jsContent string) *JSRuntime {
	runtime := &JSRuntime{
		Name:           name,
		EventLoop:      make(chan *event),
		VM:             goja.New(),
		apiFunctions:   [](APIFunction){},
		jsContent:      jsContent,
		apiInitialized: false,
		destroyed:      false,
	}
	return runtime
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
		runtimes: make(map[string](*JSRuntime)),
	}
}

// NewRuntime ...
func (jsEngine *JSEngine) NewRuntime(name string, jsContent string) (*JSRuntime, error) {
	jsEngine.ready.Lock()

	_, found := jsEngine.runtimes[name]
	if found {
		jsEngine.ready.Unlock()
		return nil, fmt.Errorf("runtime=%v already exist", name)
	}

	runtime := jsEngine.newRuntime(name, jsContent)
	jsEngine.runtimes[name] = runtime

	jsEngine.ready.Unlock()

	return runtime, nil
}

// DestroyRuntime ...
func (jsEngine *JSEngine) DestroyRuntime(name string) error {
	jsEngine.ready.Lock()

	if runtime, found := jsEngine.runtimes[name]; found {

		runtime.destroy()

		delete(jsEngine.runtimes, name)

		jsEngine.ready.Unlock()
	} else {
		jsEngine.ready.Unlock()
		return fmt.Errorf("namespace=%v not found for finishing", name)
	}
	return nil
}

// AddAPI ...
func (runtime *JSRuntime) AddAPI(apiFunction APIFunction) {
	runtime.apiFunctions = append(runtime.apiFunctions, apiFunction)
}
