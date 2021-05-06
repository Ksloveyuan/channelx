package channelx

import (
	"sync"
)

// RunInParallel can process the inputs in parallel by the specified workers
func RunInParallel(inputs []interface{}, worker func(interface{})(interface{}, error), workers int) ([]interface{}, error) {
	var (
		err         error
		outputs []interface{}
		workChannel = make(chan interface{})
		errChannel  = make(chan error, workers)
		wg          = &sync.WaitGroup{}
		mux         = sync.Mutex{}
	)

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for input := range workChannel {
				output, err:=worker(input)
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
		case err = <-errChannel:
			break loop
		}
	}

	close(workChannel)
	wg.Wait()

	// double check if there is any error happens
	if err == nil {
		select {
		case err = <-errChannel:
		default:
		}
	}

	return outputs, err
}