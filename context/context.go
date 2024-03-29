package context

import (
	"fmt"
	"sync"
)

// ParentContextAlreadyInClosingStateError ...
type ParentContextAlreadyInClosingStateError struct{}

func (err *ParentContextAlreadyInClosingStateError) Error() string {
	return fmt.Sprintf("Context already in closing state, you cannot bind childs for it")
}

// Context ...
type Context interface {
	NewContextFor(instance ContextedInstance, componentName string, componentType string) (Context, error) // create new child context
	SetOnBeforeClosing(handler func(Context))                                                              // this handler calls for current context before closing all child and subchild contexts
	Opened() chan struct{}                                                                                 // channel what closes when all childs are closed and you can close current context
	Cancel()                                                                                               // sends signal to current and all child contexts to close hierarchy gracefully (childs first, parent second)
	Log(arguments ...interface{})                                                                          // log context even
	log(objects []interface{})
	wait()
}

// ContextedInstance ...
type ContextedInstance interface {
	Go(current Context)
}

type tree struct {
	changesAllowed sync.Mutex
	closingAllowed sync.Mutex
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

	onBeforeClosing      func(current Context)
	onBeforeClosingMutex sync.Mutex

	debuggerNodePath []DebugNode // it is not a pointer, it is full array copy
	debuggerMutex    sync.Mutex
}

func newContextFor(instance ContextedInstance, debugger Debugger) (Context, error) {

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
		onBeforeClosing:       func(current Context) {},
		closed:                false,
	}

	newContext.start()

	return newContext, nil
}

func (context *ctx) Opened() chan struct{} {
	return context.opened
}

func (context *ctx) recursiveSetChildsCreatingAllowed(value bool) {
	context.Log(103, "recursiveSetChildsCreatingAllowed", "...")
	for _, child := range context.childs {
		child.recursiveSetChildsCreatingAllowed(value)
	}
	context.childsCreatingAllowed = value
	context.Log(103, "recursiveSetChildsCreatingAllowed", "done")
}

func (context *ctx) close() {
	context.Log(105, "close", "...")
	context.closedMutex.Lock()
	defer context.closedMutex.Unlock()

	if !context.closed {
		context.Log(105, "close", "channel closed")
		close(context.opened)
		context.closed = true
	}
	context.Log(105, "close", "done")
}

func (context *ctx) recursiveClosing() {
	context.Log(103, "recursiveClosing", "...")
	context.callOnCloseHandler()

	childs := make(map[int64]*ctx)

	for id, child := range context.childs {
		childs[id] = child
	}

	for id, child := range childs {
		child.recursiveClosing()
		delete(context.childs, id)
	}

	context.Log(104, "childsWaitGroup", "...")
	context.childsWaitGroup.Wait()
	context.Log(104, "childsWaitGroup", "done")

	context.close()

	context.Log(104, "loopWaitGroup", "...")
	context.loopWaitGroup.Wait()
	context.Log(104, "loopWaitGroup", "done")

	context.Log(103, "recursiveClosing", "done")
}

func (context *ctx) cancel() {
	context.tree.closingAllowed.Lock()
	defer context.tree.closingAllowed.Unlock()

	{
		context.Log(102, "cancel", "recursiveSetChildsCreatingAllowed ...")
		context.tree.changesAllowed.Lock()
		context.recursiveSetChildsCreatingAllowed(false)
		context.tree.changesAllowed.Unlock()
		context.Log(102, "cancel", "recursiveSetChildsCreatingAllowed done")
	}

	{
		context.Log(102, "cancel", "recursiveClosing ...")
		context.recursiveClosing()
		context.Log(102, "cancel", "recursiveClosing done")
	}

}

func (context *ctx) Cancel() {
	go func() {
		context.cancel()
	}()
}

// StartNewFor ...
func (context *ctx) NewContextFor(instance ContextedInstance, componentName string, componentType string) (Context, error) {

	// attach to parent new child
	parent := context

	context.tree.changesAllowed.Lock()
	defer parent.tree.changesAllowed.Unlock()

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
		onBeforeClosing:       func(current Context) {},
		closed:                false,
	}

	if parent.childsCreatingAllowed {
		parent.childs[parent.nextChildID] = newContext
		parent.nextChildID++
		parent.childsWaitGroup.Add(1)
		newContext.start()
		return newContext, nil
	}
	return nil, &ParentContextAlreadyInClosingStateError{}
}

func (context *ctx) wait() {
	context.Log(101, "waiting till childs finished")
	context.childsWaitGroup.Wait()
	context.Log(101, "waiting till loop finished")
	context.loopWaitGroup.Wait()
	context.Log(101, "waiting done")
}

func (context *ctx) Log(arguments ...interface{}) {

	objects := make([]interface{}, 0)

	for _, argument := range arguments {
		objects = append(objects, argument)
	}

	context.log(objects)
}

func (context *ctx) log(objects []interface{}) {
	context.debuggerMutex.Lock()
	context.tree.debugger.Log(context.debuggerNodePath, objects)
	context.debuggerMutex.Unlock()
}

func (context *ctx) start() {

	context.loopWaitGroup.Add(1)

	go func(ctx *ctx) {

		ctx.Log(100, "started")

		{ // wait till context execution would be finished, only after that you can dispose all context resources, otherwise it could try to create new child context on disposed resources
			ctx.instance.Go(ctx)
			ctx.Log(100, "finished")
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

func (context *ctx) SetOnBeforeClosing(handler func(current Context)) {
	context.onBeforeClosingMutex.Lock()
	defer context.onBeforeClosingMutex.Unlock()
	context.onBeforeClosing = handler
}

func (context *ctx) callOnCloseHandler() {
	context.onBeforeClosingMutex.Lock()
	defer context.onBeforeClosingMutex.Unlock()
	if context.onBeforeClosing != nil {
		context.onBeforeClosing(context)
	}
}
