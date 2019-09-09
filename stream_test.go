package channelutil

import (
	"fmt"
	"testing"
	"github.com/pkg/errors"
)

func TestChannelStream_Pipe_StopWhenHasError(t *testing.T) {
	stream := NewChannelStream(func(seedChan chan<- Result) {
		for i := 1; i <= 10; i++ {
			if i %2 == 1 {
				seedChan <- Result{Data: i, Err: nil}
			} else {
				seedChan <- Result{Err:errors.New("An error")}
			}
		}
		close(seedChan)
	}, StopWhenHasError)

	harvestResult := 0
	stream.Done(func(result Result) {
		if result.Err == nil {
			i := result.Data.(int)
			harvestResult += i
		}
	})

	fmt.Println(harvestResult)
}

func TestChannelStream_Pipe_ResumeWhenHasError(t *testing.T) {
	stream := NewChannelStream(func(seedChan chan<- Result) {
		for i := 1; i <= 10; i++ {
			if i %2 == 1 {
				seedChan <- Result{Data: i, Err: nil}
			} else {
				seedChan <- Result{Err:errors.New("An error")}
			}
		}
		close(seedChan)
	}, ResumeWhenHasError)

	harvestResult := 0
	stream.Done(func(result Result) {
		if result.Err == nil {
			i := result.Data.(int)
			harvestResult += i
		}
	})

	fmt.Println(harvestResult)
}

func TestChannelStream_Pipe_Twice(t *testing.T) {
	stream := NewChannelStream(func(seedChan chan<- Result) {
		for i := 1; i <= 100; i++ {
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
	stream3.Done(func(result Result) {
		i := result.Data.(int)
		harvestResult += i
	})

	fmt.Println(harvestResult)
}

func TestChannelStream_Pipe(t *testing.T) {
	stream := NewChannelStream(func(seedChan chan<- Result) {
		for i := 1; i <= 100; i++ {
			seedChan <- Result{Data: i, Err: nil}
		}
		close(seedChan)
	})

	stream2 := stream.Pipe(func(result Result) Result {
		return Result{Data: result.Data.(int) * 2}
	})

	harvestResult := 0
	stream2.Done(func(result Result) {
		i := result.Data.(int)
		harvestResult += i
	})

	fmt.Println(harvestResult)
}

func TestChannelStream_Done(t *testing.T) {
	stream := NewChannelStream(func(seedChan chan<- Result) {
		for i := 1; i <= 100; i++ {
			seedChan <- Result{Data: i, Err: nil}
		}
		close(seedChan)
	})

	harvestResult := 0
	stream.Done(func(result Result) {
		i := result.Data.(int)
		harvestResult += i
	})

	fmt.Println(harvestResult)
}
