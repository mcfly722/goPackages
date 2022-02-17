package scheduler_test

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
	"time"

	"github.com/mcfly722/goPackages/scheduler"
)

// TakeFirstOutdated ...
func getFirstOutdatedWithWaiting(scheduler scheduler.Scheduler) interface{} {
	for {
		object := scheduler.TakeFirstOutdatedOrNil()
		if object != nil {
			return object
		}
	}
}

func testQueueWithLenght(t *testing.T, scheduler scheduler.Scheduler, n int) {

	fmt.Print("test for:")
	now := time.Now()

	for i := 0; i < n; i++ {
		v := rand.Intn(10)
		fmt.Print(fmt.Sprintf("%v", v))
		scheduler.RegisterNewTimer(now.Add(time.Duration(v)*time.Nanosecond), v)
	}

	result := []int{}

	for i := 0; i < n; i++ {
		a := getFirstOutdatedWithWaiting(scheduler)
		result = append(result, a.(int))
	}

	fmt.Println(fmt.Sprintf("%v", result))
	if !sort.IntsAreSorted(result) {
		t.Fatal(fmt.Sprintf("obtained unsorted timers! %v", result))
	}

}

func Test_First(t *testing.T) {
	scheduler := scheduler.NewScheduler()
	scheduler.RegisterNewTimer(time.Now(), 1)
	obj := getFirstOutdatedWithWaiting(scheduler)
	if obj == nil {
		t.Fatal("no outdated objects")
	}
}

func Test_Empty(t *testing.T) {
	scheduler := scheduler.NewScheduler()
	object := scheduler.TakeFirstOutdatedOrNil()
	if object != nil {
		t.Fatal("empty scheduler returns object")
	}
}

func Test_Nby3(t *testing.T) {
	scheduler := scheduler.NewScheduler()
	for i := 0; i < 1000; i++ {
		testQueueWithLenght(t, scheduler, 3)
	}
}

func Test_CancelOne(t *testing.T) {
	scheduler := scheduler.NewScheduler()
	scheduler.RegisterNewTimer(time.Now(), 1)
	scheduler.CancelTimerFor(1)
	object := scheduler.TakeFirstOutdatedOrNil()
	if object != nil {
		t.Fatal("empty scheduler returns object")
	}
}

func Test_CancelSeveral(t *testing.T) {
	scheduler := scheduler.NewScheduler()
	scheduler.RegisterNewTimer(time.Now(), 1)
	scheduler.RegisterNewTimer(time.Now(), 1)
	scheduler.RegisterNewTimer(time.Now(), 2)
	scheduler.RegisterNewTimer(time.Now(), 1)
	scheduler.RegisterNewTimer(time.Now(), 2)
	scheduler.RegisterNewTimer(time.Now(), 2)
	scheduler.CancelTimerFor(1)
	scheduler.CancelTimerFor(2)

	object := scheduler.TakeFirstOutdatedOrNil()
	if object != nil {
		t.Fatal("empty scheduler returns object")
	}
}

func Test_CancelAll(t *testing.T) {
	scheduler := scheduler.NewScheduler()
	scheduler.RegisterNewTimer(time.Now(), 1)
	scheduler.RegisterNewTimer(time.Now(), 2)
	scheduler.CancelAllTimers()

	object := scheduler.TakeFirstOutdatedOrNil()
	if object != nil {
		t.Fatal("empty scheduler returns object")
	}
}
