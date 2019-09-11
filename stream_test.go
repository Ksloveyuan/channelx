package channelutil

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"testing"
	"time"
)

func TestChannelStream_SecondPipe_StopWhenHasError(t *testing.T) {
	ctx := context.Background()
	stream := NewChannelStream(ctx, func(seedChan chan<- Result) {
		for i := 1; i <= 10; i++ {
			time.Sleep(1 * time.Millisecond)
			if i%2 == 1 {
				seedChan <- Result{Data: i, Err: nil}
			} else {
				seedChan <- Result{Err: errors.New("An error")}
			}
		}
		close(seedChan)
	})

	stream2 := stream.Pipe(ctx, func(result Result) Result {
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
}
func TestChannelStream_Cancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	stream := NewChannelStream(ctx, func(seedChan chan<- Result) {
		for i := 1; i <= 100; i++ {
			time.Sleep(1 * time.Millisecond)
			if i%2 == 1 {
				seedChan <- Result{Data: i, Err: nil}
			} else {
				seedChan <- Result{Err: errors.New(fmt.Sprintf("error: %d", i))}
			}

			if i == 5 {
				cancel()
			}
		}
		close(seedChan)
	})

	harvestResult := 0
	_, errs := stream.Harvest(func(result Result) {
		if result.Err == nil {
			i := result.Data.(int)
			harvestResult += i
		}
	})

	fmt.Println(harvestResult)
	for _, err := range errs {
		fmt.Println("err", err)
	}
}

func TestChannelStream_Race(t *testing.T) {
	ctx := context.Background()
	stream := NewChannelStream(ctx, func(seedChan chan<- Result) {
		for i := 1; i <= 10; i++ {
			time.Sleep(1 * time.Millisecond)
			if i%2 == 1 {
				seedChan <- Result{Data: i, Err: nil}
			} else {
				seedChan <- Result{Err: errors.New(fmt.Sprintf("error: %d", i))}
			}
		}
		close(seedChan)
	})

	var error error
	stream.Race(ctx, func(result Result) bool {
		if result.Err == nil {
			fmt.Println("skip", result.Data.(int))
			return false
		}
		error = result.Err
		return true
	})

	fmt.Println(error)
}

func TestChannelStream_Harvest_StopWhenHasError(t *testing.T) {
	ctx := context.Background()
	stream := NewChannelStream(ctx, func(seedChan chan<- Result) {
		for i := 1; i <= 10; i++ {
			time.Sleep(1 * time.Millisecond)
			if i%2 == 1 {
				seedChan <- Result{Data: i, Err: nil}
			} else {
				seedChan <- Result{Err: errors.New("An error")}
			}
		}
		close(seedChan)
	}, StopWhenHasError())

	harvestResult := 0
	stream.Harvest(func(result Result) {
		if result.Err == nil {
			i := result.Data.(int)
			harvestResult += i
		}
	})

	fmt.Println(harvestResult)
}

func TestChannelStream_Harvest_ResumeWhenHasError(t *testing.T) {
	ctx := context.Background()
	stream := NewChannelStream(ctx, func(seedChan chan<- Result) {
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
	ctx := context.Background()
	stream := NewChannelStream(ctx, func(seedChan chan<- Result) {
		for i := 1; i <= 100; i++ {
			time.Sleep(1 * time.Millisecond)
			seedChan <- Result{Data: i, Err: nil}
		}
		close(seedChan)
	})

	stream2 := stream.Pipe(ctx, func(result Result) Result {
		return Result{Data: result.Data.(int) * 2}
	})

	stream3 := stream2.Pipe(ctx, func(result Result) Result {
		return Result{Data: result.Data.(int) - 1}
	})

	harvestResult := 0
	stream3.Harvest(func(result Result) {
		i := result.Data.(int)
		harvestResult += i
	})

	fmt.Println(harvestResult)
}

func TestChannelStream_Pipe(t *testing.T) {
	ctx := context.Background()
	stream := NewChannelStream(ctx, func(seedChan chan<- Result) {
		for i := 1; i <= 100; i++ {
			seedChan <- Result{Data: i, Err: nil}
		}
		close(seedChan)
	})

	stream2 := stream.Pipe(ctx, func(result Result) Result {
		return Result{Data: result.Data.(int) * 2}
	})

	harvestResult := 0
	stream2.Harvest(func(result Result) {
		i := result.Data.(int)
		harvestResult += i
	})

	fmt.Println(harvestResult)
}

func TestChannelStream_Wait(t *testing.T) {
	ctx := context.Background()
	stream := NewChannelStream(ctx, func(seedChan chan<- Result) {
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
}
