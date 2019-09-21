package channelx

import (
	"errors"
	"log"
	"sync"
	"testing"
	"time"
)

func TestAggregator_Basic(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(100)

	batchProcess := func(items []interface{}) error {
		defer wg.Add(-len(items))
		time.Sleep(120 * time.Millisecond)
		if len(items) != 4 {
			return errors.New("len(items) != 4")
		}
		return nil
	}

	errorHandler := func(err error, items []interface{}, batchProcessFunc BatchProcessFunc, aggregator *Aggregator) {
		if err == nil {
			t.FailNow()
		}
		log.Println("Receive error")
	}

	aggregator := NewAggregator(batchProcess, func(option AggregatorOption) AggregatorOption {
		option.batchSize = 4
		option.workers = 2
		option.maxWaitTime = 1 * time.Second
		option.logger = newConsoleLogger()
		option.errorHandler = errorHandler
		option.skipIfQueueIsFull = false
		return option
	})

	aggregator.Start()

	for i := 0; i < 100; i++ {
		aggregator.Enqueue(i)
		time.Sleep(10 * time.Millisecond)
	}

	wg.Wait()
	aggregator.Stop()
}
