package context

import (
	"fmt"
	"sync"
)

// Context ...
type Context interface {
	NewContextFor(instance ContextedInstance, componentName string, componentType string) Context // create new child context
	Opened() chan struct{}                                                                        // channel what closes when all childs are closed and you can close current context
	Cancel()                                                                                      // sends signal to current and all child contexts to close hierarchy gracefully (childs first, parent second)
	Log(eventType int, msg string)                                                                // log context even
	wait()
}

// ContextedInstance ...
type ContextedInstance interface {
	Go(current Context)
}

type tree struct {
	changesAllowed sync.Mutex
	debugger       Debugger
}

type ctx struct {
	id                    int64
	parent                *ctx
	childs                map[int64]*ctx
	nextChildID           int64
	instance              ContextedInstance
	childsCreatingAllowed bool
	childsWaitGroup       sync.WaitGroup
	loopWaitGroup         sync.WaitGroup
	opened                chan struct{}
	tree                  *tree

	closed      bool
	closedMutex sync.Mutex

	debuggerNodePath []DebugNode // it is not a pointer, it is full array copy
	debuggerMutex    sync.Mutex
}

func newContextFor(instance ContextedInstance, debugger Debugger) Context {

	newContext := &ctx{
		id:                    0,
		parent:                nil,
		debuggerNodePath:      []DebugNode{DebugNode{ID: 0, ComponentType: "root", ComponentName: "root"}},
		childs:                make(map[int64]*ctx),
		nextChildID:           0,
		instance:              instance,
		childsCreatingAllowed: true,
		opened:                make(chan struct{}),
		tree:                  &tree{debugger: debugger},
		closed:                false,
	}

	newContext.start()

	return newContext
}

func (context *ctx) Opened() chan struct{} {
	return context.opened
}

func (context *ctx) recursiveSetChildsCreatingAllowed(value bool) {
	for _, child := range context.childs {
		child.recursiveSetChildsCreatingAllowed(value)
	}
	context.childsCreatingAllowed = value
}

func (context *ctx) close() {
	context.closedMutex.Lock()
	defer context.closedMutex.Unlock()

	if !context.closed {
		close(context.opened)
		context.closed = true
	}
}

func (context *ctx) recursiveClosing() {
	//context.Log(101, "recursiveClosingChilds ...")
	childs := make(map[int64]*ctx)

	for id, child := range context.childs {
		childs[id] = child
	}

	for id, child := range childs {
		child.recursiveClosing()
		delete(context.childs, id)
	}

	context.childsWaitGroup.Wait()
	//context.Log(101, "recursiveClosingChilds done")

	context.close()
	context.loopWaitGroup.Wait()
}

func (context *ctx) cancel() {
	//context.Log(101, "recursiveSetChildsCreatingAllowed ...")
	context.tree.changesAllowed.Lock()
	context.recursiveSetChildsCreatingAllowed(false)
	context.tree.changesAllowed.Unlock()
	//context.Log(101, "recursiveSetChildsCreatingAllowed done")

	//context.Log(101, "recursiveClosing ...")
	context.tree.changesAllowed.Lock()
	context.recursiveClosing()
	context.tree.changesAllowed.Unlock()
	//context.Log(101, "recursiveClosing done")
}

func (context *ctx) Cancel() {
	go func() {
		context.cancel()
	}()
}

// StartNewFor ...
func (context *ctx) NewContextFor(instance ContextedInstance, componentName string, componentType string) Context {

	// attach to parent new child
	parent := context

	context.tree.changesAllowed.Lock()

	parent.debuggerMutex.Lock()
	debuggerNodePath := make([]DebugNode, len(parent.debuggerNodePath))
	copy(debuggerNodePath, parent.debuggerNodePath)
	newDebuggerNodePath := append(debuggerNodePath, DebugNode{ID: parent.nextChildID, ComponentName: componentName, ComponentType: componentType})
	parent.debuggerMutex.Unlock()

	newContext := &ctx{
		id:                    parent.nextChildID,
		parent:                parent,
		debuggerNodePath:      newDebuggerNodePath,
		childs:                make(map[int64]*ctx),
		nextChildID:           0,
		instance:              instance,
		childsCreatingAllowed: parent.childsCreatingAllowed,
		opened:                make(chan struct{}),
		tree:                  parent.tree,
		closed:                false,
	}

	if parent.childsCreatingAllowed {
		parent.childs[parent.nextChildID] = newContext
		parent.nextChildID++
		parent.childsWaitGroup.Add(1)
		newContext.start()
	}

	parent.tree.changesAllowed.Unlock()

	return newContext

}

func (context *ctx) wait() {
	context.Log(101, "waiting till finished")
	context.childsWaitGroup.Wait()
	context.loopWaitGroup.Wait()
}

func (context *ctx) Log(eventType int, msg string) {
	context.debuggerMutex.Lock()
	context.tree.debugger.Log(context.debuggerNodePath, eventType, msg)
	context.debuggerMutex.Unlock()
}

func (context *ctx) start() {

	context.loopWaitGroup.Add(1)

	go func(ctx *ctx) {

		ctx.Log(100, "started")

		{ // wait till context execution would be finished, only after that you can dispose all context resources, otherwise it could try to create new child context on disposed resources
			ctx.instance.Go(ctx)
			ctx.Log(101, "finished")
		}

		{ // panic on not closed childs
			if len(ctx.childs) != 0 {
				ctx.debuggerMutex.Lock()
				node := ctx.debuggerNodePath[len(ctx.debuggerNodePath)-1]
				panic(fmt.Sprintf("you tries to exit from context %v[%v] that have unclosed childs. Use context.Close() method, instead just exiting from goroutine!", node.ComponentType, node.ComponentName))
			}
		}

		{ // childs WaitGroup decremented
			if ctx.parent != nil {
				ctx.parent.childsWaitGroup.Done()
			}
		}

		{ // loop finished
			ctx.loopWaitGroup.Done()
		}

	}(context)
}
