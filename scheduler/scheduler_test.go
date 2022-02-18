package scheduler_test

import (
	"fmt"
	"sort"
	"strconv"
	"sync"
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

func testQueueWithLenght(t *testing.T, scheduler scheduler.Scheduler, sequence *[]int) {

	now := time.Now()

	var wg sync.WaitGroup
	for i := 0; i < len(*sequence); i++ {
		v := (*sequence)[i]
		fmt.Print(fmt.Sprintf("%v", v))
		wg.Add(1)
		go func() { // test for concurrency
			scheduler.RegisterNewTimer(now.Add(time.Duration(v)*time.Nanosecond), v)
			wg.Done()
		}()
	}
	wg.Wait()

	result := []int{}

	for i := 0; i < len(*sequence); i++ {
		a := getFirstOutdatedWithWaiting(scheduler)
		result = append(result, a.(int))
	}

	fmt.Println(fmt.Sprintf(" -> %v", result))
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

func string2Combination(str string) *[]int {
	b := []byte(str)
	out := []int{}
	for _, v := range b {
		out = append(out, int(v)-48)
	}
	return &out
}

func Test_RecombineFirstN(t *testing.T) {
	scheduler := scheduler.NewScheduler()
	for i := 1; i < 1024; i++ {
		str := strconv.FormatInt(int64(i), 4)
		sequence := string2Combination(str)
		testQueueWithLenght(t, scheduler, sequence)
	}
}
