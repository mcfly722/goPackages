package jsEngine

import (
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/mcfly722/goPackages/context"
)

// Scheduler ...
type Scheduler struct {
	context   context.Context
	eventLoop EventLoop
	runtime   *goja.Runtime
}

// Constructor ...
func (scheduler Scheduler) Constructor(context context.Context, eventLoop EventLoop, runtime *goja.Runtime) {
	runtime.Set("Scheduler", &Scheduler{
		context:   context,
		eventLoop: eventLoop,
		runtime:   runtime,
	})
}

// Ticker ...
type Ticker struct {
	scheduler  *Scheduler
	intervalMs int64
	spreadMs   int64
	handler    *goja.Callable

	ready sync.Mutex
}

// StartedTicker ...
type StartedTicker struct {
	ticker *activeTicker
}

type activeTicker struct {
	scheduler     *Scheduler
	intervalMs    int64
	spreadMs      int64
	handler       *goja.Callable
	opened        chan struct{}
	alreadyClosed bool

	ready sync.Mutex
}

// NewTicker ...
func (scheduler *Scheduler) NewTicker(intervalMs int64, handler *goja.Callable) *Ticker {
	return &Ticker{
		scheduler:  scheduler,
		intervalMs: intervalMs,
		spreadMs:   0,
		handler:    handler,
	}
}

// SetInitialSpread ...
func (ticker *Ticker) SetInitialSpread(spreadMs int64) *Ticker {
	ticker.ready.Lock()
	defer ticker.ready.Unlock()
	ticker.spreadMs = spreadMs
	return ticker
}

// Start ...
func (ticker *Ticker) Start() *StartedTicker {
	ticker.ready.Lock()
	defer ticker.ready.Unlock()

	started := &StartedTicker{
		ticker: &activeTicker{
			scheduler:     ticker.scheduler,
			intervalMs:    ticker.intervalMs,
			spreadMs:      ticker.spreadMs,
			handler:       ticker.handler,
			opened:        make(chan struct{}),
			alreadyClosed: false,
		},
	}

	_, err := ticker.scheduler.context.NewContextFor(started.ticker, "ticker", "ticker")
	if err != nil {
		panic(ticker.scheduler.runtime.ToValue(err.Error()))
	}

	return started
}

func (ticker *activeTicker) Go(current context.Context) {
	delay := time.Duration(time.Duration(ticker.spreadMs) * time.Millisecond)
loop:
	for {
		select {
		case _, opened := <-current.Opened():
			if !opened {
				break loop
			}
			break
		case _, opened := <-ticker.opened:
			if !opened {
				current.Cancel()
			}
			break
		case <-time.After(delay):
			delay = time.Duration(time.Duration(ticker.intervalMs) * time.Millisecond)
			_, err := ticker.scheduler.eventLoop.CallHandler(ticker.handler, ticker.scheduler.runtime.ToValue(nil))
			if err != nil {
				current.Log(51, err.Error())
				current.Cancel()
			}
			break
		}
	}

}

// Stop ...
func (startedTicker *StartedTicker) Stop() {
	startedTicker.ticker.ready.Lock()
	defer startedTicker.ticker.ready.Unlock()

	if !startedTicker.ticker.alreadyClosed {
		close(startedTicker.ticker.opened)
		startedTicker.ticker.alreadyClosed = true
	}
}
