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
	work    WorkFunc
	promise *Promise
}

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
			request.promise.res, request.promise.err = request.work()
			close(request.promise.done)
		case <-actor.quit:
			break loop
		}
	}
	actor.wg.Done()
}

// Do a work.
func (actor *Actor) Do(workFunc WorkFunc) *Promise {
	methodRequest := request{work: workFunc, promise: &Promise{
		done: make(chan struct{}),
	}}
	actor.queue <- methodRequest
	return methodRequest.promise
}

// Close actor
func (actor *Actor) Close() {
	close(actor.quit)
	actor.wg.Wait()
}
