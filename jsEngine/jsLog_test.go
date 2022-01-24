package jsEngine_test

import (
	"testing"

	"github.com/mcfly722/goPackages/jsEngine"
)

func Test_Log(t *testing.T) {
	engine := jsEngine.NewJSEngine()
	engine.AddAPI(jsEngine.JSLog)

	if err := engine.NewRuntime("test", "log('logger works!')"); err != nil {
		t.Fatal(err)
	}

	if err := engine.CloseRuntime("test"); err != nil {
		t.Fatal(err)
	}
}
