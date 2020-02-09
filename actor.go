package channelx

import (
	"runtime"
	"sync"
)

// Represents the actor
type Actor struct {
	buffer int
	queue  chan request

	wg   *sync.WaitGroup
	quit chan struct{}
}

// Represents a request
type request struct {
	work WorkFunc
	call *Call
}

// Represent the state of the call
type Call struct {
	done   chan struct{}
	error  error
	result interface{}
}

// Represents the func that Actor do
type WorkFunc func() (interface{}, error)

// Represents the func to set actor option
type SetActorOptionFunc func(actor *Actor)

// Set actor buffer
func SetActorBuffer(buffer int) SetActorOptionFunc {
	return func(actor *Actor) {
		actor.buffer = buffer
	}
}

// Creates a new actor
func NewActor(setActorOptionFuncs ...SetActorOptionFunc) *Actor {
	actor := &Actor{buffer: runtime.NumCPU(), quit: make(chan struct{}), wg: &sync.WaitGroup{}}
	for _, setOptionFunc := range setActorOptionFuncs {
		setOptionFunc(actor)
	}

	actor.queue = make(chan request, actor.buffer)

	actor.wg.Add(1)
	go actor.schedule()

	return actor
}

// The long live go routine to run.
func (actor *Actor) schedule() {
loop:
	for {
		select {
		case request := <-actor.queue:
			request.call.result, request.call.error = request.work()
			close(request.call.done)
		case <-actor.quit:
			break loop
		}
	}
	actor.wg.Done()
}

// Do a work.
func (actor *Actor) Do(workFunc WorkFunc) *Call {
	methodRequest := request{work: workFunc, call: &Call{
		done: make(chan struct{}),
	}}
	actor.queue <- methodRequest
	return methodRequest.call
}

// Close actor
func (actor *Actor) Close() {
	close(actor.quit)
	actor.wg.Wait()
}

// Wait for the call completes
func (call *Call) Done() (interface{}, error) {
	<-call.done
	return call.result, call.error
}
