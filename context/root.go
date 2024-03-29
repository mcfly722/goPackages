package context

// RootContext ...
type RootContext interface {
	NewContextFor(instance ContextedInstance, componentName string, componentType string) (Context, error) // create new child context
	Cancel()                                                                                               // cancel root context with all childs
	Wait()                                                                                                 // waits till root context would be closed
	Log(vars ...interface{})                                                                               // log context event
}

// Root ...
type Root struct {
	ctx Context
}

// Go ...
func (root *Root) Go(current Context) {
loop:
	for {
		select {
		case _, opened := <-current.Opened():
			if !opened {
				break loop
			}
		}
	}
}

// NewRootContext ...
func NewRootContext(debugger Debugger) RootContext {
	root := &Root{}

	root.ctx, _ = newContextFor(root, debugger)

	return root
}

// Cancel ...
func (root *Root) Cancel() {
	root.ctx.Cancel()
}

// Wait ...
func (root *Root) Wait() {
	root.ctx.wait()
}

// Log ...
func (root *Root) Log(arguments ...interface{}) {
	objects := make([]interface{}, 0)

	for _, argument := range arguments {
		objects = append(objects, argument)
	}

	root.ctx.log(objects)
}

// NewContextFor ...
func (root *Root) NewContextFor(instance ContextedInstance, componentName string, componentType string) (Context, error) {
	return root.ctx.NewContextFor(instance, componentName, componentType)
}
