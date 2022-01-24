package jsEngine

import (
	"fmt"
	"log"
)

// JSLog ...
type JSLog struct{}

// Initialize ...
func (jsLog *JSLog) Initialize(runtime *JSRuntime) error {

	log.Printf(fmt.Sprintf("api:log added to %v", runtime.Name))

	logger := func(msg string) {
		log.Println(msg)
	}

	runtime.VM.Set("log", logger)

	return nil

}

// Dispose ...
func (jsLog *JSLog) Dispose(runtime *JSRuntime) {
	log.Printf(fmt.Sprintf("api:log disposed for %v", runtime.Name))
}
