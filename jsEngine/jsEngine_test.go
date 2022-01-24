package jsEngine_test

import (
	"fmt"
	"io/ioutil"
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

func Test_Race(t *testing.T) {
	body, err := ioutil.ReadFile("jsEngine_test.js")
	if err != nil {
		t.Fatal(err)
	}

	api := func(runtime *goja.Runtime) {}

	engine1 := jsEngine.NewJSEngine(api)

	engine2 := jsEngine.NewJSEngine(api)

	t.Log(fmt.Sprintf("script:\n%v", string(body)))

	for i := 0; i < 10; i++ {
		err = engine1.NewRuntime(fmt.Sprintf("1-%v", i), string(body))
		if err != nil {
			t.Fatal(err)
		}
		engine2.NewRuntime(fmt.Sprintf("2-%v", i), string(body))
		if err != nil {
			t.Fatal(err)
		}
	}

	for i := 0; i < 10; i++ {
		err := engine1.CloseRuntime(fmt.Sprintf("1-%v", i))
		if err != nil {
			t.Fatal(err)
		}
		err = engine2.CloseRuntime(fmt.Sprintf("2-%v", i))
		if err != nil {
			t.Fatal(err)
		}
	}

}
