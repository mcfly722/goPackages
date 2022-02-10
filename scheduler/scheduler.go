package scheduler

import "time"

// Scheduler ...
type Scheduler struct {
	queue []*timer
}

type timer struct {
	deadline time.Time
	object   interface{}
}

// NewScheduler ...
func NewScheduler() *Scheduler {
	return &Scheduler{
		queue: []*timer{},
	}
}

// RegisterNewTimer ...
func (scheduler *Scheduler) RegisterNewTimer(deadline time.Time, object interface{}) {

	newTimer := &timer{
		deadline: deadline,
		object:   object,
	}

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

// TakeFirstOutdated ...
func (scheduler *Scheduler) TakeFirstOutdated() interface{} {
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
