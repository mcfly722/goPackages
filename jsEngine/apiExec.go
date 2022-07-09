package jsEngine

import (
	"bufio"
	"errors"
	"os"
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

// Command ...
type Command struct {
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
type Process struct {
	process *process
}

type process struct {
	exec                 *Exec
	command              *exec.Cmd
	expiredAt            time.Time
	exitCode             int
	finish               chan struct{}
	stdoutStrings        chan string
	stdoutStringsHandler *goja.Callable
	onDoneHandler        *goja.Callable
}

// Constructor ...
func (exec Exec) Constructor(context context.Context, eventLoop EventLoop, runtime *goja.Runtime) {
	runtime.Set("Exec", &Exec{
		context:   context,
		eventLoop: eventLoop,
		runtime:   runtime,
	})
}

// NewCommand ...
func (exec *Exec) NewCommand(name string, args []string) *Command {
	return &Command{
		exec:      exec,
		name:      name,
		args:      args,
		directory: "",
		timeout:   0,
	}
}

// SetPath ...
func (command *Command) SetPath(directory string) *Command {
	command.directory = directory
	return command
}

// SetTimeoutMs ...
func (command *Command) SetTimeoutMs(timeoutMs int64) *Command {
	command.ready.Lock()
	defer command.ready.Unlock()

	command.timeout = time.Duration(timeoutMs) * time.Millisecond
	return command
}

// SetOnDone ...
func (command *Command) SetOnDone(handler *goja.Callable) *Command {
	command.ready.Lock()
	defer command.ready.Unlock()

	command.onDoneHandler = handler
	return command
}

// SetOnStdoutString ...
func (command *Command) SetOnStdoutString(handler *goja.Callable) *Command {
	command.ready.Lock()
	defer command.ready.Unlock()

	command.stdoutStringsHandler = handler
	return command
}

// Start ...
func (command *Command) Start() *Process {

	command.ready.Lock()
	defer command.ready.Unlock()

	cmd := exec.Command(command.name, command.args...)

	cmd = setCommandParameters(cmd)

	if len(command.directory) > 0 {
		_, err := os.Stat(command.directory)
		if err != nil {
			panic(command.exec.runtime.ToValue(err.Error()))
		}
		cmd.Dir = command.directory
	}

	started := &Process{
		process: &process{
			exec:                 command.exec,
			command:              cmd,
			exitCode:             -1,
			finish:               make(chan struct{}),
			stdoutStrings:        make(chan string),
			stdoutStringsHandler: command.stdoutStringsHandler,
			onDoneHandler:        command.onDoneHandler,
		},
	}

	if command.stdoutStringsHandler != nil {

		pipe, err := cmd.StdoutPipe()
		if err != nil {
			panic(command.exec.runtime.ToValue(err.Error()))
		}

		scanner := bufio.NewScanner(pipe)
		scanner.Split(bufio.ScanLines)

		go func(scanner *bufio.Scanner, stringsOut chan string) {
			for scanner.Scan() {
				stringsOut <- scanner.Text()
			}
			close(stringsOut)
		}(scanner, started.process.stdoutStrings)
	}

	err := cmd.Start()
	if err != nil {
		panic(command.exec.runtime.ToValue(err.Error()))
	}

	if command.timeout != 0 {
		started.process.expiredAt = time.Now().Add(command.timeout)
	}

	_, err = command.exec.context.NewContextFor(started.process, command.name, "process")
	if err != nil {
		panic(command.exec.runtime.ToValue(err.Error()))
	}

	go func(process *process, command *exec.Cmd, finish chan struct{}) {
		if err := command.Wait(); err != nil {
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
	}(started.process, cmd, started.process.finish)

	return started
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

	if err := process.command.Process.Kill(); err != nil {
		if !errors.Is(err, syscall.EINVAL) {
			current.Log(50, "killing process", err.Error())
		}
	}

}

// Stop ...
func (started *Process) Stop() {
	if err := started.process.command.Process.Kill(); err != nil {
		if !errors.Is(err, syscall.EINVAL) {
			panic(started.process.exec.runtime.ToValue(err.Error()))
		}
	}
}
