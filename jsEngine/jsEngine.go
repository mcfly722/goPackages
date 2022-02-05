package jsEngine

import (
	"fmt"
	"sync"

	"github.com/dop251/goja"
	"github.com/mcfly722/goPackages/logger"
)

const (
	defaultLogSizeForEngine  = 100
	defaultLogSizeForRuntime = 100
)

// JSRuntime ...
type JSRuntime struct {
	Name           string
	EventLoop      chan *loopEvent
	apiFunctions   [](APIFunction)
	VM             *goja.Runtime
	jsContent      string
	apiInitialized bool
	destroyed      bool
	Logger         *logger.Logger
}

// CallCallback ...
func (runtime *JSRuntime) CallCallback(callback *goja.Callable, args ...goja.Value) {
	go func() { // unblocking send callback
		runtime.EventLoop <- &loopEvent{
			Type:     1,
			callback: callback,
			args:     args,
		}
	}()
}

// Finish ...
func (runtime *JSRuntime) Finish() {
	go func() { // unblocking send exit
		runtime.EventLoop <- &loopEvent{
			Type: 0,
		}
	}()
}

// CallException ...
func (runtime *JSRuntime) CallException(functionName string, message string) {
	runtime.Logger.LogEvent(logger.EventTypeTrace, runtime.Name, fmt.Sprintf("issued function exception: %v - %v", functionName, message))
	panic(runtime.VM.ToValue(message))
}

// CallHandlerException ...
func (runtime *JSRuntime) CallHandlerException(err error) {
	runtime.Logger.LogEvent(logger.EventTypeTrace, runtime.Name, fmt.Sprintf("issued handler exception: %v", err))
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

	runtime.Logger.LogEvent(logger.EventTypeInfo, runtime.Name, "runtime started")

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

				// exit
				if event.Type == 0 {
					break loop
				}

				// callback
				if event.Type == 1 {
					_, err := (*event.callback)(nil, event.args...)
					if err != nil {
						runtime.CallHandlerException(err)
					}
				}

			}
		}

		runtime.Logger.LogEvent(logger.EventTypeInfo, runtime.Name, "runtime finished")
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
		runtime.Finish()
	}

	runtime.destroyed = true
}

// JSEngine ...
type JSEngine struct {
	runtimes map[string]*JSRuntime
	ready    sync.Mutex
	Logger   *logger.Logger
}

func (jsEngine *JSEngine) newRuntime(name string, jsContent string) *JSRuntime {
	logger := logger.NewLogger(defaultLogSizeForRuntime)
	logger.SetOutputToConsole(jsEngine.Logger.IsOutputToConsoleEnabled())

	runtime := &JSRuntime{
		Name:           name,
		EventLoop:      make(chan *loopEvent),
		VM:             goja.New(),
		apiFunctions:   [](APIFunction){},
		jsContent:      jsContent,
		apiInitialized: false,
		destroyed:      false,
		Logger:         logger,
	}
	return runtime
}

// event from JSEngine to Goroutine with JS Loop
type loopEvent struct {
	Type int
	// 0 - exit (finish go-routine)
	// 1 - callback
	callback *goja.Callable
	args     []goja.Value
}

// NewJSEngine ...
func NewJSEngine() *JSEngine {
	logger := logger.NewLogger(defaultLogSizeForEngine)

	return &JSEngine{
		runtimes: make(map[string](*JSRuntime)),
		Logger:   logger,
	}
}

// NewRuntime ...
func (jsEngine *JSEngine) NewRuntime(name string, jsContent string, logSize int) (*JSRuntime, error) {
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
