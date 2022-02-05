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
	numberOftimer int64
	timer         map[int64]*timer
}

type timer struct {
	id                   int64
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

func newTimer(runtime *JSRuntime, id int64, intervalMilliseconds int64, callback *goja.Callable) *timer {

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

	scheduler.timer = make(map[int64]*timer)
	scheduler.numberOftimer = 0

	setInterval := func(callback goja.Callable, intervalMilliseconds goja.Value) int64 {

		if intervalMilliseconds.ToInteger() < 1 {
			runtime.CallException("setInterval", fmt.Sprintf("interval should be > 0 (obtained %v)", intervalMilliseconds.ToInteger()))
			return -1
		}

		timerID := scheduler.numberOftimer
		scheduler.timer[timerID] = newTimer(runtime, scheduler.numberOftimer, intervalMilliseconds.ToInteger(), &callback)
		scheduler.numberOftimer++
		return timerID
	}

	clearInterval := func(timerID goja.Value) {
		if timer, ok := scheduler.timer[timerID.ToInteger()]; ok {

			timer.finish()
			delete(scheduler.timer, timerID.ToInteger())

		} else {
			runtime.CallException("clearInterval", fmt.Sprintf("timer with Id=%v not found", timerID))
		}
	}

	runtime.VM.Set("setInterval", setInterval)
	runtime.VM.Set("clearInterval", clearInterval)

	log.Printf(fmt.Sprintf("api:scheduler initialized for %v", runtime.Name))

	return nil

}

// Dispose ...
func (scheduler *JSScheduler) Dispose(runtime *JSRuntime) {

	for _, timer := range scheduler.timer {
		timer.finish()
	}

	scheduler.timer = make(map[int64]*timer)
	scheduler.numberOftimer = 0

	log.Printf(fmt.Sprintf("api:scheduler disposed for %v", runtime.Name))
}
