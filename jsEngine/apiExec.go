package jsEngine

import (
	"bufio"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/dop251/goja"
	"github.com/mcfly722/goPackages/context"
)

// Exec ...
type Exec struct {
	context   context.Context
	eventLoop EventLoop
	runtime   *goja.Runtime
}

// Cmd ...
type Cmd struct {
	exec                 *Exec
	name                 string
	args                 []string
	directory            string
	timeout              time.Duration
	onDoneHandler        *goja.Callable
	stdoutStringsHandler *goja.Callable

	ready sync.Mutex
}

// Process ...
type Process struct{}

type process struct {
	exec                 *Exec
	expiredAt            time.Time
	exitCode             int
	finish               chan struct{}
	stdoutStrings        chan string
	stdoutStringsHandler *goja.Callable
	onDoneHandler        *goja.Callable
	//	eventLoop            EventLoop
	//runtime              *goja.Runtime
}

// Constructor ...
func (exec Exec) Constructor(context context.Context, eventLoop EventLoop, runtime *goja.Runtime) {
	//executer := runtime.NewObject()
	//executer.Set("NewCommand", NewCommand)
	runtime.Set("Exec", &Exec{
		context:   context,
		eventLoop: eventLoop,
		runtime:   runtime,
	})
}

// NewCommand ...
func (exec *Exec) NewCommand(name string, args []string) *Cmd {
	return &Cmd{
		exec:      exec,
		name:      name,
		args:      args,
		directory: "",
		timeout:   0,
	}
}

// SetPath ...
func (cmd *Cmd) SetPath(directory string) *Cmd {
	cmd.directory = directory
	return cmd
}

// SetTimeoutMs ...
func (cmd *Cmd) SetTimeoutMs(timeoutMs int64) *Cmd {
	cmd.ready.Lock()
	defer cmd.ready.Unlock()

	cmd.timeout = time.Duration(timeoutMs) * time.Millisecond
	return cmd
}

// SetOnDone ...
func (cmd *Cmd) SetOnDone(handler *goja.Callable) *Cmd {
	cmd.ready.Lock()
	defer cmd.ready.Unlock()

	cmd.onDoneHandler = handler
	return cmd
}

// SetOnStdString ...
func (cmd *Cmd) SetOnStdoutString(handler *goja.Callable) *Cmd {
	cmd.ready.Lock()
	defer cmd.ready.Unlock()

	cmd.stdoutStringsHandler = handler
	return cmd
}

// Start ...
func (cmd *Cmd) Start() *Process {

	cmd.ready.Lock()
	defer cmd.ready.Unlock()

	command := exec.Command(cmd.name, cmd.args...)

	if len(cmd.directory) > 0 {
		command.Dir = cmd.directory
	}

	proc := &process{
		exec:                 cmd.exec,
		exitCode:             -1,
		finish:               make(chan struct{}),
		stdoutStrings:        make(chan string),
		stdoutStringsHandler: cmd.stdoutStringsHandler,
		onDoneHandler:        cmd.onDoneHandler,
		//		eventLoop:            cmd.eventLoop,
		//		runtime:              cmd.runtime,
	}

	if cmd.stdoutStringsHandler != nil {

		pipe, err := command.StdoutPipe()
		if err != nil {
			panic(cmd.exec.runtime.ToValue(err.Error()))
		}

		scanner := bufio.NewScanner(pipe)
		scanner.Split(bufio.ScanLines)

		go func(scanner *bufio.Scanner, stringsOut chan string) {
			for scanner.Scan() {
				stringsOut <- scanner.Text()
			}
			close(stringsOut)
		}(scanner, proc.stdoutStrings)
	}

	err := command.Start()
	if err != nil {
		panic(cmd.exec.runtime.ToValue(err.Error()))
	}

	if cmd.timeout != 0 {
		proc.expiredAt = time.Now().Add(cmd.timeout)
	}

	_, err = cmd.exec.context.NewContextFor(proc, cmd.name, "process")
	if err != nil {
		panic(cmd.exec.runtime.ToValue(err.Error()))
	}

	go func(process *process, cmd *exec.Cmd, finish chan struct{}) {
		if err := cmd.Wait(); err != nil {
			if exiterr, ok := err.(*exec.ExitError); ok {
				// The program has exited with an exit code != 0

				// This works on both Unix and Windows. Although package
				// syscall is generally platform dependent, WaitStatus is
				// defined for both Unix and Windows and in both cases has
				// an ExitStatus() method with the same signature.
				if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
					process.exitCode = status.ExitStatus()
				}
			}
		}
		close(finish)
	}(proc, command, proc.finish)

	startedProcess := &Process{}

	return startedProcess
}

// Go ...
func (process *process) Go(current context.Context) {

loop:
	for {
		select {
		case _, opened := <-current.Opened():
			if !opened {
				break loop
			}
			break
		case _, opened := <-process.finish:
			if !opened {
				current.Cancel()
			}
			break
		case stdoutString, opened := <-process.stdoutStrings:
			if opened {
				if process.stdoutStringsHandler != nil {
					process.exec.eventLoop.CallHandler(process.stdoutStringsHandler, process.exec.runtime.ToValue(stdoutString))
				}
			}
			break

		default:
			if !process.expiredAt.IsZero() {
				if time.Now().After(process.expiredAt) {
					process.expiredAt = time.Time{} // to do not call several cancels on expiration

					current.Log(50, "timeouted")
					current.Cancel()
				}
			}
		}
	}

	if process.onDoneHandler != nil {
		process.exec.eventLoop.CallHandler(process.onDoneHandler, process.exec.runtime.ToValue(process.exitCode))
	}

}
