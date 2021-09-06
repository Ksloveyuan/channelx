package channelx_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Ksloveyuan/channelx"
)

func TestChannelStream_SecondPipe_StopWhenHasError(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	stream := channelx.NewChannelStream(func(seedChan chan<- channelx.Item, quitChannel chan struct{}) {
		defer wg.Done()
	loop:
		for i := 1; i <= 10; i++ {
			time.Sleep(1 * time.Millisecond)
			fmt.Println("seed", i)
			var result channelx.Item
			if i%2 == 1 {
				result = channelx.Item{Data: i, Err: nil}

			} else {
				result = channelx.Item{Err: errors.New(fmt.Sprintf("error %d", i))}
			}

			select {
			case <-quitChannel:
				break loop
			case seedChan <- result:
			}
		}
		close(seedChan)
		fmt.Println("seed finished")
	}, channelx.SetWorkers(1))

	stream2 := stream.Pipe(func(result channelx.Item) channelx.Item {
		if result.Err == nil {
			i := result.Data.(int) * 2
			fmt.Println("s2", i)
			return channelx.Item{Data: i}
		}

		return channelx.Item{Err: result.Err}
	}, channelx.StopWhenHasError(), channelx.SetWorkers(1))

	harvestResult := 0
	stream2.Harvest(func(result channelx.Item) {
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
	stream := channelx.NewChannelStream(func(seedChan chan<- channelx.Item, quitChannel chan struct{}) {
		defer wg.Done()
	loop:
		for i := 1; i <= 10; i++ {
			time.Sleep(1 * time.Millisecond)
			fmt.Println("seed", i)
			var result channelx.Item
			if i%2 == 1 {
				result = channelx.Item{Data: i, Err: nil}

			} else {
				result = channelx.Item{Err: errors.New(fmt.Sprintf("error %d", i))}
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
	_, errs := stream.Harvest(func(result channelx.Item) {
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
	stream := channelx.NewChannelStream(func(seedChan chan<- channelx.Item, quitChannel chan struct{}) {
		defer wg.Done()
	loop:
		for i := 1; i <= 10; i++ {
			time.Sleep(1 * time.Millisecond)
			fmt.Println("seed", i)
			var result channelx.Item
			if i%2 == 1 {
				result = channelx.Item{Data: i, Err: nil}

			} else {
				result = channelx.Item{Err: errors.New(fmt.Sprintf("error %d", i))}
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
	_, errs := stream.Harvest(func(result channelx.Item) {
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
	stream := channelx.NewChannelStream(func(seedChan chan<- channelx.Item, quitChannel chan struct{}) {
		defer wg.Done()
	loop:
		for i := 1; i <= 10; i++ {
			fmt.Println("seed", i)
			var result channelx.Item
			if i%2 == 1 {
				result = channelx.Item{Data: i, Err: nil}
			} else {
				result = channelx.Item{Err: errors.New(fmt.Sprintf("error %d", i))}
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
	}, channelx.StopWhenHasError(), channelx.SetWorkers(1))

	var error error
	stream.Race(func(result channelx.Item) bool {
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
	stream := channelx.NewChannelStream(func(seedChan chan<- channelx.Item, quitChannel chan struct{}) {
		defer wg.Done()
	loop:
		for i := 1; i <= 10; i++ {
			time.Sleep(1 * time.Millisecond)
			fmt.Println("seed", i)
			var result channelx.Item
			if i%2 == 1 {
				result = channelx.Item{Data: i, Err: nil}

			} else {
				result = channelx.Item{Err: errors.New(fmt.Sprintf("error %d", i))}
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
	}, channelx.StopWhenHasError(), channelx.SetWorkers(1))

	harvestResult := 0
	stream.Harvest(func(result channelx.Item) {
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
	stream := channelx.NewChannelStream(func(seedChan chan<- channelx.Item, quitChannel chan struct{}) {
		for i := 1; i <= 10; i++ {
			time.Sleep(1 * time.Millisecond)
			if i%2 == 1 {
				seedChan <- channelx.Item{Data: i, Err: nil}
			} else {
				seedChan <- channelx.Item{Err: errors.New("An error")}
			}
		}
		close(seedChan)
	}, channelx.ResumeWhenHasError())

	harvestResult := 0
	stream.Harvest(func(result channelx.Item) {
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
	stream := channelx.NewChannelStream(func(seedChan chan<- channelx.Item, quitChannel chan struct{}) {
		defer wg.Done()
		for i := 1; i <= 100; i++ {
			time.Sleep(1 * time.Millisecond)
			seedChan <- channelx.Item{Data: i, Err: nil}
		}
		close(seedChan)
	})

	stream2 := stream.Pipe(func(result channelx.Item) channelx.Item {
		return channelx.Item{Data: result.Data.(int) * 2}
	})

	stream3 := stream2.Pipe(func(result channelx.Item) channelx.Item {
		return channelx.Item{Data: result.Data.(int) - 1}
	})

	harvestResult := 0
	stream3.Harvest(func(result channelx.Item) {
		i := result.Data.(int)
		harvestResult += i
	})

	fmt.Println(harvestResult)
	wg.Wait()
}

func TestChannelStream_Pipe(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	stream := channelx.NewChannelStream(func(seedChan chan<- channelx.Item, quitChannel chan struct{}) {
		defer wg.Done()
		for i := 1; i <= 100; i++ {
			seedChan <- channelx.Item{Data: i, Err: nil}
		}
		close(seedChan)
	})

	stream2 := stream.Pipe(func(result channelx.Item) channelx.Item {
		return channelx.Item{Data: result.Data.(int) * 2}
	})

	harvestResult := 0
	stream2.Harvest(func(result channelx.Item) {
		i := result.Data.(int)
		harvestResult += i
	})

	fmt.Println(harvestResult)
	wg.Wait()
}

func TestChannelStream(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	stream := channelx.NewChannelStream(func(seedChan chan<- channelx.Item, quitChannel chan struct{}) {
		defer wg.Done()
		for i := 1; i <= 100; i++ {
			seedChan <- channelx.Item{Data: i, Err: nil}
		}
		close(seedChan)
	})

	harvestResult := 0
	stream.Harvest(func(result channelx.Item) {
		i := result.Data.(int)
		harvestResult += i
	})

	fmt.Println(harvestResult)
	wg.Wait()
}
