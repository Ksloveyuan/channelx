package channelx

import (
	"runtime"
	"sync"
	"time"
)

// Represents the aggregator
type Aggregator struct {
	option         AggregatorOption
	wg             *sync.WaitGroup
	quit           chan struct{}
	eventQueue     chan interface{}
	batchProcessor BatchProcessFunc
}

// Represents the aggregator option
type AggregatorOption struct {
	batchSize         int
	workers           int
	maxWaitTime       time.Duration
	errorHandler      ErrorHandlerFunc
	skipIfQueueIsFull bool
	logger            Logger
}

// the func to batch process items
type BatchProcessFunc func([]interface{}) error

// the func to set option for aggregator
type SetOptionFunc func(option AggregatorOption) AggregatorOption

// the func to handle error
type ErrorHandlerFunc func(err error, items []interface{}, batchProcessFunc BatchProcessFunc, aggregator *Aggregator)

// Creates a new aggregator
func NewAggregator(batchProcessor BatchProcessFunc, optionFuncs ...SetOptionFunc) *Aggregator {
	option := AggregatorOption{
		batchSize:         32,
		workers:           runtime.NumCPU(),
		maxWaitTime:       1 * time.Minute,
		skipIfQueueIsFull: true,
	}

	for _, optionFunc := range optionFuncs {
		option = optionFunc(option)
	}

	return &Aggregator{
		eventQueue:     make(chan interface{}, 2*option.workers),
		option:         option,
		quit:           make(chan struct{}),
		wg:             new(sync.WaitGroup),
		batchProcessor: batchProcessor,
	}
}

// Enqueue an item
func (agt *Aggregator) Enqueue(item interface{}) bool {
	select {
	case agt.eventQueue <- item:
		return true
	default:
		if agt.option.logger != nil {
			agt.option.logger.Warnc("Aggregator", nil, "Event queue is full")
		}

		runtime.Gosched()

		select {
		case agt.eventQueue <- item:
			return true
		default:
			if !agt.option.skipIfQueueIsFull {
				if agt.option.logger != nil {
					agt.option.logger.Warnc("Aggregator", nil, "Async enqueue event %+v", item)
				}
				go agt.Enqueue(item) // this must not be the best way to handle it
			} else if agt.option.logger != nil {
				agt.option.logger.Warnc("Aggregator", nil, "Event queue is still full and %+v is skipped.", item)
			}
			return false
		}
	}
}

// Start the aggregator
func (agt *Aggregator) Start() {
	for i := 0; i < agt.option.workers; i++ {
		index := i
		go agt.work(index)
	}
}

// Stop the aggregator
func (agt *Aggregator) Stop() {
	close(agt.quit)
	agt.wg.Wait()
}

// Stop the aggregator safely, the difference with Stop is it guarantees no item is missed during stop
func (agt *Aggregator) SafeStop() {
	if len(agt.eventQueue) == 0 {
		close(agt.quit)
	} else {
		ticker := time.NewTicker(10 * time.Millisecond)
		for range ticker.C {
			if len(agt.eventQueue) == 0 {
				close(agt.quit)
				break
			}
		}
		ticker.Stop()
	}
	agt.wg.Wait()
}

func (agt *Aggregator) work(index int) {
	defer func() {
		if r := recover(); r != nil {
			agt.work(index)
		}
	}()

	agt.wg.Add(1)
	defer agt.wg.Done()

	reqs := make([]interface{}, 0, agt.option.batchSize)
	idleDelay := time.NewTimer(agt.option.maxWaitTime)
	defer idleDelay.Stop()

loop:
	for {
		idleDelay.Reset(agt.option.maxWaitTime)
		select {
		case req := <-agt.eventQueue:
			reqs = append(reqs, req)
			if len(reqs) < agt.option.batchSize {
				break
			}

			agt.batchProcess(reqs)
			reqs = make([]interface{}, 0, agt.option.batchSize)
		case <-idleDelay.C:
			if len(reqs) == 0 {
				break
			}

			agt.batchProcess(reqs)
			reqs = make([]interface{}, 0, agt.option.batchSize)
		case <-agt.quit:
			break loop
		}
	}
}

func (agt *Aggregator) batchProcess(items []interface{}) {
	if err := agt.batchProcessor(items); err != nil {
		if agt.option.logger != nil {
			agt.option.logger.Errorc("Aggregator", err, "error happens")
		}

		if agt.option.errorHandler != nil {
			go agt.option.errorHandler(err, items, agt.batchProcessor, agt)
		} else if agt.option.logger != nil {
			agt.option.logger.Errorc("Aggregator", err, "error happens in batchProcess and is skipped")
		}
	} else if agt.option.logger != nil {
		agt.option.logger.Infoc("Aggregator", "%d items have been sent.", len(items))
	}
}
