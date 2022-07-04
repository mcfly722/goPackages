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
type Process struct{}

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
func (cmd *Command) SetPath(directory string) *Command {
	cmd.directory = directory
	return cmd
}

// SetTimeoutMs ...
func (cmd *Command) SetTimeoutMs(timeoutMs int64) *Command {
	cmd.ready.Lock()
	defer cmd.ready.Unlock()

	cmd.timeout = time.Duration(timeoutMs) * time.Millisecond
	return cmd
}

// SetOnDone ...
func (cmd *Command) SetOnDone(handler *goja.Callable) *Command {
	cmd.ready.Lock()
	defer cmd.ready.Unlock()

	cmd.onDoneHandler = handler
	return cmd
}

// SetOnStdoutString ...
func (cmd *Command) SetOnStdoutString(handler *goja.Callable) *Command {
	cmd.ready.Lock()
	defer cmd.ready.Unlock()

	cmd.stdoutStringsHandler = handler
	return cmd
}

// StartNewProcess ...
func (cmd *Command) StartNewProcess() *Process {

	cmd.ready.Lock()
	defer cmd.ready.Unlock()

	command := exec.Command(cmd.name, cmd.args...)

	command = setCommandParameters(command)

	if len(cmd.directory) > 0 {
		_, err := os.Stat(cmd.directory)
		if err != nil {
			panic(cmd.exec.runtime.ToValue(err.Error()))
		}
		command.Dir = cmd.directory
	}

	proc := &process{
		exec:                 cmd.exec,
		command:              command,
		exitCode:             -1,
		finish:               make(chan struct{}),
		stdoutStrings:        make(chan string),
		stdoutStringsHandler: cmd.stdoutStringsHandler,
		onDoneHandler:        cmd.onDoneHandler,
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

	if err := process.command.Process.Kill(); err != nil {
		if !errors.Is(err, syscall.EINVAL) {
			current.Log(50, "killing process", err.Error())
		}
	}

}
