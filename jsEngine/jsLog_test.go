package jsEngine_test

import (
	"testing"

	"github.com/mcfly722/goPackages/jsEngine"
)

func Test_Log(t *testing.T) {
	engine := jsEngine.NewJSEngine()

	runtime, err := engine.NewRuntime("test", "log('logger works!')", 0)
	if err != nil {
		t.Fatal(err)
	}

	runtime.AddAPI(&jsEngine.JSLog{})

	if err := runtime.Start(); err != nil {
		t.Fatal(err)
	}

	if err := engine.DestroyRuntime("test"); err != nil {
		t.Fatal(err)
	}
}
