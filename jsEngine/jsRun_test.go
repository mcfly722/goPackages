package jsEngine_test

import (
	"testing"
	"time"

	"github.com/mcfly722/goPackages/jsEngine"
)

func Test_Run(t *testing.T) {
	engine := jsEngine.NewJSEngine()

	script := `
run('cmd.exe', ['/c','dir'], 1000, function(output){
  log(output)
})
  `

	runtime, err := engine.NewRuntime("test", script, 0)
	if err != nil {
		t.Fatal(err)
	}

	runtime.AddAPI(&jsEngine.JSRun{})
	runtime.AddAPI(&jsEngine.JSLog{})

	if err := runtime.Start(); err != nil {
		t.Fatal(err)
	}

	time.Sleep(1 * time.Second)

	if err := engine.DestroyRuntime("test"); err != nil {
		t.Fatal(err)
	}
}
