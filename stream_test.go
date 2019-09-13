package channelx

import (
	"fmt"
	"github.com/pkg/errors"
	"sync"
	"testing"
	"time"
)

func TestChannelStream_SecondPipe_StopWhenHasError(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	stream := NewChannelStream(func(seedChan chan<- Result, quitChannel chan struct{}) {
		defer wg.Done()
	loop:
		for i := 1; i <= 10; i++ {
			time.Sleep(1 * time.Millisecond)
			fmt.Println("seed", i)
			var result Result
			if i%2 == 1 {
				result = Result{Data: i, Err: nil}

			} else {
				result = Result{Err: errors.New(fmt.Sprintf("error %d", i))}
			}

			select {
			case <-quitChannel:
				break loop
			case seedChan <- result:
			}
		}
		close(seedChan)
		fmt.Println("seed finished")
	}, SetWorkers(1))

	stream2 := stream.Pipe(func(result Result) Result {
		if result.Err == nil {
			i := result.Data.(int) * 2
			fmt.Println("s2", i)
			return Result{Data: i}
		}

		return Result{Err: result.Err}
	}, StopWhenHasError(), SetWorkers(1))

	harvestResult := 0
	stream2.Harvest(func(result Result) {
		if result.Err == nil {
			i := result.Data.(int)
			fmt.Println("har", i)
			harvestResult += i
		} else {
			fmt.Println("met error")
		}
	})

	fmt.Println(harvestResult)
	wg.Wait()
}

func TestChannelStream_Cancel_ByCloseExplicitly(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	stream := NewChannelStream(func(seedChan chan<- Result, quitChannel chan struct{}) {
		defer wg.Done()
	loop:
		for i := 1; i <= 10; i++ {
			time.Sleep(1 * time.Millisecond)
			fmt.Println("seed", i)
			var result Result
			if i%2 == 1 {
				result = Result{Data: i, Err: nil}

			} else {
				result = Result{Err: errors.New(fmt.Sprintf("error %d", i))}
			}

			select {
			case <-quitChannel:
				break loop
			case seedChan <- result:
			}

			if i == 5 {
				close(quitChannel)
			}
		}
		close(seedChan)
		fmt.Println("seed finished")
	})

	harvestResult := 0
	_, errs := stream.Harvest(func(result Result) {
		fmt.Println("harvest", result)
		if result.Err == nil {
			if i, ok := result.Data.(int); ok {
				harvestResult += i
			}
		}
	})

	fmt.Println(harvestResult)
	for _, err := range errs {
		fmt.Println("err", err)
	}

	wg.Wait()
}

func TestChannelStream_Cancel(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	stream := NewChannelStream(func(seedChan chan<- Result, quitChannel chan struct{}) {
		defer wg.Done()
	loop:
		for i := 1; i <= 10; i++ {
			time.Sleep(1 * time.Millisecond)
			fmt.Println("seed", i)
			var result Result
			if i%2 == 1 {
				result = Result{Data: i, Err: nil}

			} else {
				result = Result{Err: errors.New(fmt.Sprintf("error %d", i))}
			}

			select {
			case <-quitChannel:
				break loop
			case seedChan <- result:
			}
		}
		close(seedChan)
		fmt.Println("seed finished")
	})

	harvestResult := 0
	_, errs := stream.Harvest(func(result Result) {
		fmt.Println("harvest", result)
		if result.Err == nil {
			if i, ok := result.Data.(int); ok {
				harvestResult += i
			}
		}

		if harvestResult > 2 {
			stream.Cancel()
		}
	})

	fmt.Println(harvestResult)
	for _, err := range errs {
		fmt.Println("err", err)
	}

	wg.Wait()
}

func TestChannelStream_Race(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	stream := NewChannelStream(func(seedChan chan<- Result, quitChannel chan struct{}) {
		defer wg.Done()
	loop:
		for i := 1; i <= 10; i++ {
			fmt.Println("seed", i)
			var result Result
			if i%2 == 1 {
				result = Result{Data: i, Err: nil}
			} else {
				result = Result{Err: errors.New(fmt.Sprintf("error %d", i))}
			}

			select {
			case <-quitChannel:
				fmt.Println("break seed loop")
				break loop
			case seedChan <- result:
			}
		}
		close(seedChan)
		fmt.Println("seed finished")
	}, StopWhenHasError(), SetWorkers(1))

	var error error
	stream.Race(func(result Result) bool {
		if result.Err == nil {
			fmt.Println("skip", result.Data.(int))
			return false
		}
		error = result.Err
		return true
	})

	fmt.Println(error)
	wg.Wait()
}

func TestChannelStream_Harvest_StopWhenHasError(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	stream := NewChannelStream(func(seedChan chan<- Result, quitChannel chan struct{}) {
		defer wg.Done()
	loop:
		for i := 1; i <= 10; i++ {
			time.Sleep(1 * time.Millisecond)
			fmt.Println("seed", i)
			var result Result
			if i%2 == 1 {
				result = Result{Data: i, Err: nil}

			} else {
				result = Result{Err: errors.New(fmt.Sprintf("error %d", i))}
			}

			select {
			case <-quitChannel:
				fmt.Println("break seed loop")
				break loop
			case seedChan <- result:
			}
		}
		close(seedChan)
		fmt.Println("seed finished")
	}, StopWhenHasError(), SetWorkers(1))

	harvestResult := 0
	stream.Harvest(func(result Result) {
		if result.Err == nil {
			fmt.Println("harvest", result)
			if i, ok := result.Data.(int); ok {
				harvestResult += i
			} else {
				fmt.Println("harvest a strange result", result)
			}
		}
	})

	fmt.Println(harvestResult)

	wg.Wait()
}

func TestChannelStream_Harvest_ResumeWhenHasError(t *testing.T) {
	stream := NewChannelStream(func(seedChan chan<- Result, quitChannel chan struct{}) {
		for i := 1; i <= 10; i++ {
			time.Sleep(1 * time.Millisecond)
			if i%2 == 1 {
				seedChan <- Result{Data: i, Err: nil}
			} else {
				seedChan <- Result{Err: errors.New("An error")}
			}
		}
		close(seedChan)
	}, ResumeWhenHasError())

	harvestResult := 0
	stream.Harvest(func(result Result) {
		if result.Err == nil {
			i := result.Data.(int)
			harvestResult += i
		} else {
			harvestResult -= 1
		}
	})

	fmt.Println(harvestResult)
}

func TestChannelStream_Pipe_Twice(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	stream := NewChannelStream(func(seedChan chan<- Result, quitChannel chan struct{}) {
		defer wg.Done()
		for i := 1; i <= 100; i++ {
			time.Sleep(1 * time.Millisecond)
			seedChan <- Result{Data: i, Err: nil}
		}
		close(seedChan)
	})

	stream2 := stream.Pipe(func(result Result) Result {
		return Result{Data: result.Data.(int) * 2}
	})

	stream3 := stream2.Pipe(func(result Result) Result {
		return Result{Data: result.Data.(int) - 1}
	})

	harvestResult := 0
	stream3.Harvest(func(result Result) {
		i := result.Data.(int)
		harvestResult += i
	})

	fmt.Println(harvestResult)
	wg.Wait()
}

func TestChannelStream_Pipe(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	stream := NewChannelStream(func(seedChan chan<- Result, quitChannel chan struct{}) {
		defer wg.Done()
		for i := 1; i <= 100; i++ {
			seedChan <- Result{Data: i, Err: nil}
		}
		close(seedChan)
	})

	stream2 := stream.Pipe(func(result Result) Result {
		return Result{Data: result.Data.(int) * 2}
	})

	harvestResult := 0
	stream2.Harvest(func(result Result) {
		i := result.Data.(int)
		harvestResult += i
	})

	fmt.Println(harvestResult)
	wg.Wait()
}

func TestChannelStream(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	stream := NewChannelStream(func(seedChan chan<- Result, quitChannel chan struct{}) {
		defer wg.Done()
		for i := 1; i <= 100; i++ {
			seedChan <- Result{Data: i, Err: nil}
		}
		close(seedChan)
	})

	harvestResult := 0
	stream.Harvest(func(result Result) {
		i := result.Data.(int)
		harvestResult += i
	})

	fmt.Println(harvestResult)
	wg.Wait()
}
