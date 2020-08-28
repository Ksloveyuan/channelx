package channelx

import (
	"runtime"
	"sync"
)

type evalMode string

const (
	SimpleMode   evalMode = "simple"
	FutureMode   evalMode = "future"
	CallbackMode evalMode = "callback"
)

// Represent the workload function
type WorkloadFunc func() (interface{}, error)

// Represents the ActiveObject
type ActiveObject struct {
	activationQueue chan methodRequest
	waitGroup       *sync.WaitGroup
	quit            chan struct{}
}

type Future struct {
	done chan struct{}
	resp interface{}
	err  error
}

// Represents a methodRequest
type methodRequest struct {
	workload WorkloadFunc
	evalMode evalMode
	future   *Future
	callback func(interface{}, error)
}

// Wait until the future is done and get response.
func (f *Future) Done() (interface{}, error) {
	<-f.done
	return f.resp, f.err
}

func NewActiveObject() *ActiveObject {
	return NewActiveObjectWithBuffer(runtime.NumCPU())
}

// Creates a new activeObject
func NewActiveObjectWithBuffer(bufferSize int) *ActiveObject {
	activeObject := &ActiveObject{quit: make(chan struct{}), waitGroup: &sync.WaitGroup{}}

	activeObject.activationQueue = make(chan methodRequest, bufferSize)
	activeObject.waitGroup.Add(1)
	go activeObject.scheduleLoop()

	return activeObject
}

// The long live go routine to run.
func (activeObject *ActiveObject) scheduleLoop() {
loop:
	for {
		select {
		case methodRequest := <-activeObject.activationQueue:
			switch methodRequest.evalMode {
			case SimpleMode:
				_, _ = methodRequest.workload()
			case CallbackMode:
				evalRes, err := methodRequest.workload()
				methodRequest.callback(evalRes, err)
			case FutureMode:
				methodRequest.future.resp, methodRequest.future.err = methodRequest.workload()
				close(methodRequest.future.done)
			}
		case <-activeObject.quit:
			break loop
		}
	}
	activeObject.waitGroup.Done()
}

// Close activeObject
func (activeObject *ActiveObject) Close() {
	close(activeObject.quit)
	activeObject.waitGroup.Wait()
}

// Call a workload.
func (activeObject *ActiveObject) Call(workloadFunc WorkloadFunc) {
	methodRequest := methodRequest{
		workload: workloadFunc,
		evalMode: SimpleMode,
	}
	activeObject.activationQueue <- methodRequest
}

// CallWithFuture a workload.
func (activeObject *ActiveObject) CallWithFuture(workloadFunc WorkloadFunc) *Future {
	methodRequest := methodRequest{
		workload: workloadFunc,
		evalMode: FutureMode,
		future: &Future{
			done: make(chan struct{}),
		}}
	activeObject.activationQueue <- methodRequest
	return methodRequest.future
}

// CallWithCallback a workload.
func (activeObject *ActiveObject) CallWithCallback(workloadFunc WorkloadFunc, callback func(interface{}, error)) {
	methodRequest := methodRequest{
		workload: workloadFunc,
		evalMode: CallbackMode,
		callback: callback,
	}
	activeObject.activationQueue <- methodRequest
}
