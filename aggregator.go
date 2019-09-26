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
	BatchSize         int
	Workers           int
	ChannelBufferSize int
	MaxIdleTime       time.Duration
	ErrorHandler      ErrorHandlerFunc
	Logger            Logger
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
		BatchSize:   32,
		Workers:     runtime.NumCPU(),
		MaxIdleTime: 1 * time.Minute,
	}

	for _, optionFunc := range optionFuncs {
		option = optionFunc(option)
	}

	if option.ChannelBufferSize <= option.Workers {
		option.ChannelBufferSize = option.Workers
	}

	return &Aggregator{
		eventQueue:     make(chan interface{}, option.ChannelBufferSize),
		option:         option,
		quit:           make(chan struct{}),
		wg:             new(sync.WaitGroup),
		batchProcessor: batchProcessor,
	}
}

// Try enqueue an item, and it is non-blocked
func (agt *Aggregator) TryEnqueue(item interface{}) bool {
	select {
	case agt.eventQueue <- item:
		return true
	default:
		if agt.option.Logger != nil {
			agt.option.Logger.Warnc("Aggregator", nil, "Event queue is full and try reschedule")
		}

		runtime.Gosched()

		select {
		case agt.eventQueue <- item:
			return true
		default:
			if agt.option.Logger != nil {
				agt.option.Logger.Warnc("Aggregator", nil, "Event queue is still full and %+v is skipped.", item)
			}
			return false
		}
	}
}

// Enqueue an item, will be blocked if the queue is full
func (agt *Aggregator) Enqueue(item interface{}) {
	agt.eventQueue <- item
}

// Start the aggregator
func (agt *Aggregator) Start() {
	for i := 0; i < agt.option.Workers; i++ {
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
			if agt.option.Logger != nil {
				agt.option.Logger.Errorc("Aggregator", nil, "recover worker as bad thing happens %+v", r)
			}

			agt.work(index)
		}
	}()

	agt.wg.Add(1)
	defer agt.wg.Done()

	reqs := make([]interface{}, 0, agt.option.BatchSize)
	idleDelay := time.NewTimer(agt.option.MaxIdleTime)
	defer idleDelay.Stop()

loop:
	for {
		idleDelay.Reset(agt.option.MaxIdleTime)
		select {
		case req := <-agt.eventQueue:
			reqs = append(reqs, req)
			if len(reqs) < agt.option.BatchSize {
				break
			}

			agt.wg.Add(1)
			agt.batchProcess(reqs)
			agt.wg.Done()
			reqs = make([]interface{}, 0, agt.option.BatchSize)
		case <-idleDelay.C:
			if len(reqs) == 0 {
				break
			}

			agt.wg.Add(1)
			agt.batchProcess(reqs)
			agt.wg.Done()
			reqs = make([]interface{}, 0, agt.option.BatchSize)
		case <-agt.quit:
			if len(reqs) != 0 {
				agt.wg.Add(1)
				agt.batchProcess(reqs)
				agt.wg.Done()
			}

			break loop
		}
	}
}

func (agt *Aggregator) batchProcess(items []interface{}) {
	if err := agt.batchProcessor(items); err != nil {
		if agt.option.Logger != nil {
			agt.option.Logger.Errorc("Aggregator", err, "error happens")
		}

		if agt.option.ErrorHandler != nil {
			go agt.option.ErrorHandler(err, items, agt.batchProcessor, agt)
		} else if agt.option.Logger != nil {
			agt.option.Logger.Errorc("Aggregator", err, "error happens in batchProcess and is skipped")
		}
	} else if agt.option.Logger != nil {
		agt.option.Logger.Infoc("Aggregator", "%d items have been sent.", len(items))
	}
}
