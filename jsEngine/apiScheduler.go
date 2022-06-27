package jsEngine

import (
	"fmt"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/mcfly722/goPackages/context"
)

type timer struct {
	id        int64
	delayMS   int64
	scheduler *scheduler
	handler   *goja.Callable
}

// Go ...
func (timer *timer) Go(current context.Context) {
	delay := time.Duration(0)
loop:
	for {
		select {
		case <-time.After(delay):
			delay = time.Duration(time.Duration(timer.delayMS) * time.Millisecond)
			_, err := timer.scheduler.eventLoop.Call(timer.handler)
			if err != nil {
				current.Log(40, err.Error())
			}
			break
		case _, opened := <-current.Opened():
			if !opened {
				break loop
			}
		}
	}

	timer.scheduler.deleteTimer(timer.id)

}

type scheduler struct {
	timers        map[int64]*timer
	timersCounter int64
	eventLoop     *eventLoop
	ready         sync.Mutex
}

func (scheduler *scheduler) addTimer(handler *goja.Callable, delayMS int64) *timer {
	scheduler.ready.Lock()
	defer scheduler.ready.Unlock()

	timer := &timer{
		id:        scheduler.timersCounter,
		delayMS:   delayMS,
		scheduler: scheduler,
		handler:   handler,
	}

	scheduler.timers[timer.id] = timer
	scheduler.timersCounter++
	return timer
}

func (scheduler *scheduler) deleteTimer(timerID int64) error {
	scheduler.ready.Lock()
	defer scheduler.ready.Unlock()

	if _, ok := scheduler.timers[timerID]; ok {
		delete(scheduler.timers, timerID)
		return nil
	}

	return fmt.Errorf("there are no timers with id=%v", timerID)
}

func apiScheduler(context context.Context, eventLoop *eventLoop, runtime *goja.Runtime) {

	scheduler := &scheduler{
		timers:        make(map[int64]*timer, 0),
		timersCounter: 0,
		eventLoop:     eventLoop,
	}

	setInterval := func(handler *goja.Callable, delayMS int64) int64 {

		timer := scheduler.addTimer(handler, delayMS)

		_, err := context.NewContextFor(timer, fmt.Sprintf("timer%v", timer.id), "intervalTimer")
		if err != nil {
			context.Log("NewContextFor", "skipping")
			scheduler.deleteTimer(timer.id)
			return -1
		}

		return timer.id
	}

	runtime.Set("setInterval", setInterval)

}
