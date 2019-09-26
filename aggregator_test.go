package channelx

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestAggregator_Basic(t *testing.T) {
	batchProcess := func(items []interface{}) error {
		t.Logf("handler %d items", len(items))
		return nil
	}

	aggregator := NewAggregator(batchProcess)

	aggregator.Start()

	for i := 0; i < 1000; i++ {
		aggregator.TryEnqueue(i)
	}

	aggregator.SafeStop()
}

func TestAggregator_Complex(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(100)

	batchProcess := func(items []interface{}) error {
		defer wg.Add(-len(items))
		time.Sleep(200 * time.Millisecond)
		if len(items) != 4 {
			return errors.New("len(items) != 4")
		}
		return nil
	}

	errorHandler := func(err error, items []interface{}, batchProcessFunc BatchProcessFunc, aggregator *Aggregator) {
		if err == nil {
			t.FailNow()
		}
		t.Log("Receive error")
	}

	aggregator := NewAggregator(batchProcess, func(option AggregatorOption) AggregatorOption {
		option.BatchSize = 4
		option.Workers = 2
		option.MaxIdleTime = 8 * time.Millisecond
		option.Logger = NewConsoleLogger()
		option.ErrorHandler = errorHandler
		return option
	})

	aggregator.Start()

	for i := 0; i < 100; i++ {
		for !aggregator.TryEnqueue(i) {
			time.Sleep(10 * time.Millisecond)
		}
		time.Sleep(5 * time.Millisecond)
	}

	aggregator.SafeStop()
	wg.Wait()
}

func TestAggregator_IdleTimeOut(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(100)

	batchProcess := func(items []interface{}) error {
		defer wg.Add(-len(items))
		time.Sleep(50 * time.Millisecond)
		if len(items) != 4 {
			t.Log("meet max time out")
		}
		return nil
	}

	aggregator := NewAggregator(batchProcess, func(option AggregatorOption) AggregatorOption {
		option.BatchSize = 4
		option.Workers = 1
		option.MaxIdleTime = 95 * time.Millisecond
		option.Logger = NewConsoleLogger()
		return option
	})

	aggregator.Start()

	for i := 0; i < 100; i++ {
		aggregator.TryEnqueue(i)
		time.Sleep(100 * time.Millisecond)
	}

	aggregator.SafeStop()
	wg.Wait()
}
