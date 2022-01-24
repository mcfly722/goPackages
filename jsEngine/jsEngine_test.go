package jsEngine_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/mcfly722/goPackages/jsEngine"
)

type FakeAPISuccess struct{}

func (_ *FakeAPISuccess) Initialize(runtime *jsEngine.JSRuntime) error {
	fmt.Println(fmt.Sprintf("FakeAPISuccess for %v initialized", runtime.Name))
	return nil
}
func (_ *FakeAPISuccess) Dispose(runtime *jsEngine.JSRuntime) {
	fmt.Println(fmt.Sprintf("FakeAPISuccess for %v disposed", runtime.Name))
}

type fakeAPIFailed struct{}

func (_ *fakeAPIFailed) Initialize(runtime *jsEngine.JSRuntime) error {
	fmt.Println(fmt.Sprintf("fakeAPIFailed for %v initialized", runtime.Name))
	return errors.New("exception")
}
func (_ *fakeAPIFailed) Dispose(runtime *jsEngine.JSRuntime) {
	fmt.Println(fmt.Sprintf("fakeAPIFailed for %v disposed", runtime.Name))
}

func Test_JSEngine(t *testing.T) {
	engine := jsEngine.NewJSEngine()

	runtime, err := engine.NewRuntime("test", "var a = [1,2,3,4,5]")
	if err != nil {
		t.Fatal(err)
	}

	if err := runtime.Start(); err != nil {
		t.Fatal(err)
	}

	time.Sleep(1000 * time.Millisecond)
	engine.DestroyRuntime("test")
}

func Test_WrongJSCode(t *testing.T) {
	engine := jsEngine.NewJSEngine()

	runtime, err := engine.NewRuntime("test", "fuck")
	if err != nil {
		t.Fatal(err)
	}

	if err := runtime.Start(); err != nil {
		t.Log(fmt.Sprintf("%v error catched.ok.", err))
	} else {
		t.Fatal("error was not dropped")
	}

}

func Test_AppendRuntimeWithSameName(t *testing.T) {
	engine := jsEngine.NewJSEngine()

	_, err := engine.NewRuntime("test", "var a=1")
	if err != nil {
		t.Fatal(err)
	}

	_, err = engine.NewRuntime("test", "var b=1")
	if err != nil {
		t.Log(fmt.Sprintf("%v error catched.ok.", err))
	} else {
		t.Fatal("error was not dropped")
	}
}

func Test_CloseUnknownRuntime(t *testing.T) {
	engine := jsEngine.NewJSEngine()

	_, err := engine.NewRuntime("test", "var a=1")
	if err != nil {
		t.Fatal(err)
	}

	if err := engine.DestroyRuntime("test1"); err != nil {
		t.Log(fmt.Sprintf("%v error catched.ok.", err))
	} else {
		t.Fatal("error was not dropped")
	}
}

func Test_StartDestroyedRuntime(t *testing.T) {
	engine := jsEngine.NewJSEngine()

	runtime, err := engine.NewRuntime("test", "var a=1")
	if err != nil {
		t.Fatal(err)
	}

	runtime.Start()
	if err != nil {
		t.Fatal(err)
	}

	if err := engine.DestroyRuntime("test"); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Start(); err != nil {
		t.Log(fmt.Sprintf("%v error catched.ok.", err))
	} else {
		t.Fatal("error: started already destroyed runtime")

	}
}

func Test_StartRuntimeTwoTimes(t *testing.T) {
	engine := jsEngine.NewJSEngine()

	runtime, err := engine.NewRuntime("test", "var a=1")
	if err != nil {
		t.Fatal(err)
	}

	runtime.Start()
	if err != nil {
		t.Fatal(err)
	}

	runtime.Start()
	if err != nil {
		t.Fatal(err)
	}

	if err := engine.DestroyRuntime("test"); err != nil {
		t.Fatal(err)
	}

}

func Test_RollbackFailedAPI(t *testing.T) {

	engine := jsEngine.NewJSEngine()

	runtime, err := engine.NewRuntime("test", "var a = [1,2,3,4,5]")
	if err != nil {
		t.Fatal(err)
	}

	runtime.AddAPI(&FakeAPISuccess{})
	runtime.AddAPI(&fakeAPIFailed{})

	if err := runtime.Start(); err != nil {
		t.Log(fmt.Sprintf("%v successfully catched", err))
	} else {
		t.Fatal("error not catched!")
	}

	if err := runtime.Start(); err != nil {
		t.Log(fmt.Sprintf("%v successfully catched", err))
	} else {
		t.Fatal("error not catched!")
	}

	if err := engine.DestroyRuntime("test"); err != nil {
		t.Fatal(err)
	}
}

func Test_Race(t *testing.T) {
	body, err := ioutil.ReadFile("jsEngine_test.js")
	if err != nil {
		t.Fatal(err)
	}

	engine1 := jsEngine.NewJSEngine()

	engine2 := jsEngine.NewJSEngine()

	t.Log(fmt.Sprintf("script:\n%v", string(body)))

	for i := 0; i < 10; i++ {
		runtime1, err := engine1.NewRuntime(fmt.Sprintf("1-%v", i), string(body))
		if err != nil {
			t.Fatal(err)
		}

		if err := runtime1.Start(); err != nil {
			t.Fatal(err)
		}

		runtime2, err := engine2.NewRuntime(fmt.Sprintf("2-%v", i), string(body))
		if err != nil {
			t.Fatal(err)
		}

		if err := runtime2.Start(); err != nil {
			t.Fatal(err)
		}

	}

	for i := 0; i < 10; i++ {
		err := engine1.DestroyRuntime(fmt.Sprintf("1-%v", i))
		if err != nil {
			t.Fatal(err)
		}
		err = engine2.DestroyRuntime(fmt.Sprintf("2-%v", i))
		if err != nil {
			t.Fatal(err)
		}
	}

}
