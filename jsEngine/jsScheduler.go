package jsEngine

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/dop251/goja"
)

// JSScheduler ...
type JSScheduler struct {
	numberOftimer int
	timer         map[int]*timer
}

type timer struct {
	id                   int
	stop                 chan struct{}
	intervalMilliseconds int64
	nextExpectingTime    time.Time
	callback             *goja.Callable
	working              bool

	ready sync.Mutex
}

func (timer *timer) finish() {
	if timer.working {
		timer.stop <- struct{}{}
	}
}

func newTimer(runtime *JSRuntime, id int, intervalMilliseconds int64, callback *goja.Callable) *timer {

	timer := &timer{
		id:                   id,
		stop:                 make(chan struct{}),
		intervalMilliseconds: intervalMilliseconds,
		nextExpectingTime:    time.Now().Add(time.Millisecond * time.Duration(intervalMilliseconds)),
		callback:             callback,
		working:              true,
	}

	go func() {
	loop:
		for {
			select {
			case <-timer.stop:
				break loop
			default:
				if time.Now().After(timer.nextExpectingTime) {
					// fire event

					//log.Printf(fmt.Sprintf("%v timer", timer.id))
					runtime.CallCallback(callback)

					// for next time
					timer.nextExpectingTime = timer.nextExpectingTime.Add(time.Millisecond * time.Duration(intervalMilliseconds))
				}
			}
		}

		timer.working = false

		log.Printf(fmt.Sprintf("timer %v stopped", timer.id))
	}()

	log.Printf(fmt.Sprintf("timer %v started", timer.id))
	return timer
}

// Initialize ...
func (scheduler *JSScheduler) Initialize(runtime *JSRuntime) error {

	scheduler.timer = make(map[int]*timer)
	scheduler.numberOftimer = 0

	setInterval := func(callback goja.Callable, intervalMilliseconds goja.Value) {

		scheduler.timer[scheduler.numberOftimer] = newTimer(runtime, scheduler.numberOftimer, intervalMilliseconds.ToInteger(), &callback)
		scheduler.numberOftimer++
	}

	runtime.VM.Set("setInterval", setInterval)
	log.Printf(fmt.Sprintf("api:scheduler initialized for %v", runtime.Name))

	return nil

}

// Dispose ...
func (scheduler *JSScheduler) Dispose(runtime *JSRuntime) {

	for _, timer := range scheduler.timer {
		timer.finish()
	}

	scheduler.timer = make(map[int]*timer)
	scheduler.numberOftimer = 0

	log.Printf(fmt.Sprintf("api:scheduler disposed for %v", runtime.Name))
}
