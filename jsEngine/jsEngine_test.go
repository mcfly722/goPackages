package jsEngine_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/dop251/goja"
	"github.com/mcfly722/goPackages/jsEngine"
)

func Test_JSEngine1(t *testing.T) {

	api := func(runtime *goja.Runtime) {}

	engine := jsEngine.NewJSEngine(api)

	engine.NewRuntime("test", "var a = [1,2,3,4,5]")

	time.Sleep(1000 * time.Millisecond)

	engine.CloseRuntime("test")
}

func Test_WrongJSCode(t *testing.T) {
	api := func(runtime *goja.Runtime) {}
	engine := jsEngine.NewJSEngine(api)
	err := engine.NewRuntime("test", "fuck")
	if err != nil {
		t.Log(fmt.Sprintf("%v error catched.ok.", err))
	} else {
		t.Fatal("error was not dropped")
	}
}

func Test_AppendRuntimeWithSameName(t *testing.T) {
	api := func(runtime *goja.Runtime) {}
	engine := jsEngine.NewJSEngine(api)

	engine.NewRuntime("test", "var a=1")

	err := engine.NewRuntime("test", "var b=1")
	if err != nil {
		t.Log(fmt.Sprintf("%v error catched.ok.", err))
	} else {
		t.Fatal("error was not dropped")
	}
}

func Test_CloseUnknownRuntime(t *testing.T) {
	api := func(runtime *goja.Runtime) {}
	engine := jsEngine.NewJSEngine(api)

	engine.NewRuntime("test", "var a=1")

	err := engine.CloseRuntime("test1")
	if err != nil {
		t.Log(fmt.Sprintf("%v error catched.ok.", err))
	} else {
		t.Fatal("error was not dropped")
	}
}
