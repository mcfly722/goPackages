package scheduler

import (
	"sync"
	"time"
)

// Object ...
type Object interface{} // it would be nice to use generics to limit types for same interface without type assertions

// Scheduler ...
type Scheduler interface {
	RegisterNewTimer(deadline time.Time, object Object)
	TakeFirstOutdatedOrNil() Object
	CancelTimerFor(Object)
	CancelAllTimers()
}

// NewScheduler ...
func NewScheduler() Scheduler {
	return &scheduler{
		queue: []*timer{},
	}
}

// Scheduler ...
type scheduler struct {
	queue []*timer
	ready sync.Mutex
}

type timer struct {
	deadline time.Time
	object   interface{}
}

// RegisterNewTimer ...
func (scheduler *scheduler) RegisterNewTimer(deadline time.Time, object Object) {

	newTimer := &timer{
		deadline: deadline,
		object:   object,
	}

	scheduler.ready.Lock()
	defer scheduler.ready.Unlock()

	// first insert
	if len(scheduler.queue) == 0 {
		scheduler.queue = append(scheduler.queue, newTimer)
		return
	}

	// find and insert timer in correct place in queue (first element it is nearest deadline)
	// Number of operations = O(n) where n-current number of elements in sorted slice. For big slices this part could be optimized to log(n).

	insertionPos := len(scheduler.queue)
	for ; insertionPos > 0; insertionPos-- {
		if insertionPos > 0 {
			if scheduler.queue[insertionPos-1].deadline.Before(deadline) {
				break
			}
		}
	}

	if insertionPos == 0 {
		scheduler.queue = append([]*timer{newTimer}, scheduler.queue...)
		return
	}

	if insertionPos == len(scheduler.queue) {
		scheduler.queue = append(scheduler.queue, newTimer)
		return
	}

	scheduler.queue = append(scheduler.queue[:insertionPos+1], scheduler.queue[insertionPos:]...)
	scheduler.queue[insertionPos] = newTimer
}

// AsyncTakeFirstOutdated ...
func (scheduler *scheduler) TakeFirstOutdatedOrNil() Object {

	scheduler.ready.Lock()
	defer scheduler.ready.Unlock()

	if len(scheduler.queue) == 0 {
		return nil // no deadlines in queue
	}

	if scheduler.queue[0].deadline.After(time.Now()) {
		return nil // nearest deadline not outdated
	}

	// first deadline outdated
	object := scheduler.queue[0].object

	scheduler.queue = scheduler.queue[1:] // remove outdated element from queue

	return object
}

func (scheduler *scheduler) CancelTimerFor(object Object) {
	scheduler.ready.Lock()
	defer scheduler.ready.Unlock()

	// This deletion several objects from schedule.queue is not optimized enough.
	// There is another way to delete all objects for single pass, but need additional accuracy.
	for {
		lastFoundedObjectPosition := -1

		// find first occurence
	nextPass:
		for i, timer := range scheduler.queue {
			if timer.object == object {
				lastFoundedObjectPosition = i
				break nextPass
			}
		}

		// if no more object instances found in queue we just exit
		if lastFoundedObjectPosition == -1 {
			break
		} else {
			// deleting object from queue in position lastFoundedObjectIndex
			scheduler.queue = append(scheduler.queue[:lastFoundedObjectPosition], scheduler.queue[lastFoundedObjectPosition+1:]...)
		}
	}

}

// CancelAllTimers ...
func (scheduler *scheduler) CancelAllTimers() {
	scheduler.ready.Lock()
	scheduler.queue = []*timer{}
	scheduler.ready.Unlock()
}
