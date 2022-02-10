package scheduler

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func (scheduler *Scheduler) showQueue() {

	for i, timer := range scheduler.queue {
		fmt.Println(fmt.Sprintf("[%v] %v", i, timer.object))
	}

}

func Test_First(t *testing.T) {
	scheduler := NewScheduler()
	scheduler.RegisterNewTimer(time.Now(), 1)
	scheduler.showQueue()
}

func Test_FirstN(t *testing.T) {
	scheduler := NewScheduler()
	scheduler.RegisterNewTimer(time.Now().Add(1*time.Second), 1)
	scheduler.RegisterNewTimer(time.Now().Add(1*time.Second), 1)
	scheduler.RegisterNewTimer(time.Now().Add(2*time.Second), 2)
	scheduler.RegisterNewTimer(time.Now().Add(3*time.Second), 3)
	scheduler.showQueue()
}

func Test_ReverseFirstN(t *testing.T) {
	scheduler := NewScheduler()
	scheduler.RegisterNewTimer(time.Now().Add(3*time.Second), 3)
	scheduler.RegisterNewTimer(time.Now().Add(3*time.Second), 3)
	scheduler.RegisterNewTimer(time.Now().Add(2*time.Second), 2)
	scheduler.RegisterNewTimer(time.Now().Add(1*time.Second), 1)
	scheduler.showQueue()
}

func Test_InTheMiddle(t *testing.T) {
	scheduler := NewScheduler()
	scheduler.RegisterNewTimer(time.Now().Add(1*time.Second), 1)
	scheduler.RegisterNewTimer(time.Now().Add(4*time.Second), 4)
	scheduler.RegisterNewTimer(time.Now().Add(3*time.Second), 3)
	scheduler.RegisterNewTimer(time.Now().Add(2*time.Second), 2)
	scheduler.showQueue()
}

func Test_EmptyQueue(t *testing.T) {
	scheduler := NewScheduler()
	number := scheduler.TakeFirstOutdated()
	if number != nil {
		t.Fatal("empty queue returned not nul")
	}
}

func Test_AllTogetherWithRandomTimer(t *testing.T) {
	initialTime := time.Now()
	scheduler := NewScheduler()
	for i := 0; i < 100; i++ {
		delayMs := rand.Intn(1000)
		scheduler.RegisterNewTimer(initialTime.Add(time.Duration(delayMs)*time.Millisecond), delayMs)
	}

	for i := 0; i < 100; {
		number := scheduler.TakeFirstOutdated()
		if number != nil {
			i++
		}
	}

	scheduler.showQueue()
}
