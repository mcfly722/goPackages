package jsEngine_test

import (
	"testing"
	"time"

	"github.com/mcfly722/goPackages/jsEngine"
)

func Test_SchedulerSetInterval(t *testing.T) {
	engine := jsEngine.NewJSEngine()

	script := `

var timers=[];
var counters=[0,0];

timers[0] = setInterval(
	function() {
		counters[0]++
		log('timer0='+counters[0])
	},10)


timers[1] = setInterval(
	function() {
		counters[1]++

		log('timer1='+counters[1])

		if (counters[1] > 4) {
				clearInterval(timers[1])
		}

	},30)

log('timers='+timers)

	`

	runtime, err := engine.NewRuntime("test", script, 0)
	if err != nil {
		t.Fatal(err)
	}

	runtime.AddAPI(&jsEngine.JSLog{})
	runtime.AddAPI(&jsEngine.JSScheduler{})

	if err := runtime.Start(); err != nil {
		t.Fatal(err)
	}

	time.Sleep(1 * time.Second)

	if err := engine.DestroyRuntime("test"); err != nil {
		t.Fatal(err)
	}
}

func Test_SchedulerUnknownTimer(t *testing.T) {
	engine := jsEngine.NewJSEngine()

	script := `

	try {
		clearInterval(666)
	} catch(e) {
		log("catched error:"+e)
	}

	`
	runtime, err := engine.NewRuntime("test", script, 0)
	if err != nil {
		t.Fatal(err)
	}

	runtime.AddAPI(&jsEngine.JSLog{})
	runtime.AddAPI(&jsEngine.JSScheduler{})

	if err := runtime.Start(); err != nil {
		t.Fatal(err)
	}

	time.Sleep(1 * time.Second)

	if err := engine.DestroyRuntime("test"); err != nil {
		t.Fatal(err)
	}

}

func Test_SchedulerNegativeIntervalError(t *testing.T) {
	engine := jsEngine.NewJSEngine()

	script := `
	timer = setInterval(
		function() {
			log(".")
		},0)
	`
	runtime, err := engine.NewRuntime("test", script, 0)
	if err != nil {
		t.Fatal(err)
	}

	runtime.AddAPI(&jsEngine.JSLog{})
	runtime.AddAPI(&jsEngine.JSScheduler{})

	if err := runtime.Start(); err == nil {
		t.Fatal("not catched exception")
	}

	time.Sleep(1 * time.Second)

	if err := engine.DestroyRuntime("test"); err != nil {
		t.Fatal(err)
	}

}

func Test_SchedulerExceptionInsideHandler(t *testing.T) {
	engine := jsEngine.NewJSEngine()

	script := `

	count = 0
	setInterval(
		function() {
			log('call#'+count+ " started")
			count++
			throw("!!!!!!");
			log('call'+count+ " finished")
		},10)

	`
	runtime, err := engine.NewRuntime("test", script, 0)
	if err != nil {
		t.Fatal(err)
	}

	runtime.AddAPI(&jsEngine.JSLog{})
	runtime.AddAPI(&jsEngine.JSScheduler{})

	if err := runtime.Start(); err != nil {
		t.Fatal(err)
	}

	time.Sleep(1 * time.Second)

	if err := engine.DestroyRuntime("test"); err != nil {
		t.Fatal(err)
	}

}
