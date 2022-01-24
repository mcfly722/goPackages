package jsEngine_test

import (
	"testing"
	"time"

	"github.com/mcfly722/goPackages/jsEngine"
)

func Test_Scheduler(t *testing.T) {
	engine := jsEngine.NewJSEngine()

	runtime, err := engine.NewRuntime("test", "setInterval(function() {log('1')},1000)\nsetInterval(function() {log('3')},300)")
	if err != nil {
		t.Fatal(err)
	}

	runtime.AddAPI(&jsEngine.JSLog{})
	runtime.AddAPI(&jsEngine.JSScheduler{})

	if err := runtime.Start(); err != nil {
		t.Fatal(err)
	}

	time.Sleep(10 * time.Second)

	if err := engine.DestroyRuntime("test"); err != nil {
		t.Fatal(err)
	}
}
