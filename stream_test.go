package channelutil

import (
	"fmt"
	"testing"
	"github.com/pkg/errors"
)

func TestChannelStream_SecondPipe_StopWhenHasError(t *testing.T) {
	stream := NewChannelStream(func(seedChan chan<- Result) {
		for i := 1; i <= 10; i++ {
			if i %2 == 1 {
				seedChan <- Result{Data: i, Err: nil}
			} else {
				seedChan <- Result{Err:errors.New("An error")}
			}
		}
		close(seedChan)
	})

	stream2 := stream.Pipe(func(result Result) Result {
		if result.Err == nil{
			i:= result.Data.(int) * 2
			fmt.Println("s2", i)
			return Result{Data: i}
		}

		return Result{Err:result.Err}
	}, StopWhenHasError(), SetWorkers(2))

	harvestResult := 0
	stream2.Wait(func(result Result) {
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

func TestChannelStream_Race(t *testing.T) {
	stream := NewChannelStream(func(seedChan chan<- Result) {
		for i := 1; i <= 10; i++ {
			if i %2 == 1 {
				seedChan <- Result{Data: i, Err: nil}
			} else {
				seedChan <- Result{Err:errors.New(fmt.Sprintf("error: %d", i))}
			}
		}
		close(seedChan)
	})

	var error error
	stream.Race(func(result Result) bool{
		if result.Err == nil {
			fmt.Println("skip", result.Data.(int))
			return false
		}
		error = result.Err
		return true
	})

	fmt.Println(error)
}

func TestChannelStream_Wait_StopWhenHasError(t *testing.T) {
	stream := NewChannelStream(func(seedChan chan<- Result) {
		for i := 1; i <= 10; i++ {
			if i %2 == 1 {
				seedChan <- Result{Data: i, Err: nil}
			} else {
				seedChan <- Result{Err:errors.New("An error")}
			}
		}
		close(seedChan)
	}, StopWhenHasError())

	harvestResult := 0
	stream.Wait(func(result Result) {
		if result.Err == nil {
			i := result.Data.(int)
			harvestResult += i
		}
	})

	fmt.Println(harvestResult)
}

func TestChannelStream_Wait_ResumeWhenHasError(t *testing.T) {
	stream := NewChannelStream(func(seedChan chan<- Result) {
		for i := 1; i <= 10; i++ {
			if i %2 == 1 {
				seedChan <- Result{Data: i, Err: nil}
			} else {
				seedChan <- Result{Err:errors.New("An error")}
			}
		}
		close(seedChan)
	}, ResumeWhenHasError())

	harvestResult := 0
	stream.Wait(func(result Result) {
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
	stream3.Wait(func(result Result) {
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
	stream2.Wait(func(result Result) {
		i := result.Data.(int)
		harvestResult += i
	})

	fmt.Println(harvestResult)
}

func TestChannelStream_Wait(t *testing.T) {
	stream := NewChannelStream(func(seedChan chan<- Result) {
		for i := 1; i <= 100; i++ {
			seedChan <- Result{Data: i, Err: nil}
		}
		close(seedChan)
	})

	harvestResult := 0
	stream.Wait(func(result Result) {
		i := result.Data.(int)
		harvestResult += i
	})

	fmt.Println(harvestResult)
}
