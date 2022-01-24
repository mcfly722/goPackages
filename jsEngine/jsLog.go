package jsEngine

import "log"

// JSLog ...
func JSLog(jsRuntime *JSRuntime) error {

	fConsole := func(msg string) {
		log.Println(msg)
	}

	jsRuntime.VM.Set("log", fConsole)

	return nil
}
