package channelx

import (
	"context"
	"sync"
)

type quitMethod int

const (
	quitOnError         quitMethod = iota
	quitWhenAllFinished quitMethod = 1
)

// ParallelRunner represents the runner configurations
type ParallelRunner struct {
	workers    int
	quitMethod quitMethod
}

// SetParallelRunner represents the runner configuration setter
type SetParallelRunner func(runner *ParallelRunner)

// NewParallelRunner creates a new ParallelRunner
func NewParallelRunner(workers int, configurations ...SetParallelRunner) *ParallelRunner {
	runner := &ParallelRunner{workers: workers}
	for _, config := range configurations {
		config(runner)
	}
	return runner
}

// QuitOnError set runner quit on error
func QuitOnError() SetParallelRunner {
	return func(runner *ParallelRunner) {
		runner.quitMethod = quitOnError
	}
}

// QuitWhenAllFinished sets runner quit when all request are finished
func QuitWhenAllFinished() SetParallelRunner {
	return func(runner *ParallelRunner) {
		runner.quitMethod = quitWhenAllFinished
	}
}

// Run will process the requests in parallel
func (pr ParallelRunner) Run(ctx context.Context, inputs []interface{}, worker func(context.Context, interface{}) (interface{}, error)) ([]interface{}, []error) {
	var (
		outputs     []interface{}
		errors      []error
		workChannel = make(chan interface{})
		errChannel  = make(chan error, pr.workers)
		wg          = &sync.WaitGroup{}
		mux         = sync.Mutex{}
	)

	cancelContext, cancel := context.WithCancel(ctx)
	defer cancel()

	wg.Add(pr.workers)
	for i := 0; i < pr.workers; i++ {
		go func() {
			defer wg.Done()
			for input := range workChannel {
				output, err := worker(cancelContext, input)
				if err != nil {
					errChannel <- err
					break
				}

				mux.Lock()
				outputs = append(outputs, output)
				mux.Unlock()
			}
		}()
	}

loop:
	for _, input := range inputs {
		select {
		case workChannel <- input:
		case err := <-errChannel:
			errors = append(errors, err)
			if pr.quitMethod == quitOnError {
				cancel()
				break loop
			}
		}
	}

	close(workChannel)
	wg.Wait()

	// double check if there is any error happens
	if errors == nil || pr.quitMethod == quitWhenAllFinished {
		select {
		case err := <-errChannel:
			errors = append(errors, err)
		default:
		}
	}

	return outputs, errors
}

// RunInParallel is the short cut for ParallelRunner's Run
func RunInParallel(ctx context.Context, inputs []interface{}, worker func(context.Context, interface{}) (interface{}, error), workers int) ([]interface{}, error) {
	runner := NewParallelRunner(workers, QuitOnError())
	output, errors := runner.Run(ctx, inputs, worker)
	if len(errors) != 0 {
		return nil, errors[0]
	}

	return output, nil
}
