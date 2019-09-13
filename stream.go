package channelx

import (
	"runtime"
	"sync"
	"time"
)

// Represent the channel stream
type ChannelStream struct {
	dataChannel chan Item
	workers     int
	ape         actionPerError
	optionFuncs []OptionFunc
	hasError    bool
	errors      []error
	quitChan    chan struct{}
}

// Represent the item in the stream
type Item struct {
	Data interface{}
	Err  error
}

type actionPerError int

const (
	resume actionPerError = 0
	stop                  = 1
)

// The func to generate the seed in NewChannelStream
type SeedFunc func(seedChan chan<- Item, quitChannel chan struct{})

// The func to work in Pipe
type PipeFunc func(item Item) Item

// The func to harvest in Harvest
type HarvestFunc func(item Item)

// The func as a condition in Race
type RaceFunc func(item Item) bool

// The func to set option in NewChannelStream/Pipe
type OptionFunc func(cs *ChannelStream)

// Create a new channel stream
func NewChannelStream(seedFunc SeedFunc, optionFuncs ...OptionFunc) *ChannelStream {
	cs := &ChannelStream{
		workers:     runtime.NumCPU(),
		optionFuncs: optionFuncs,
	}

	for _, of := range optionFuncs {
		of(cs)
	}

	if cs.quitChan == nil {
		cs.quitChan = make(chan struct{})
	}

	cs.dataChannel = make(chan Item, cs.workers)

	go func() {
		inputChan := make(chan Item)

		go seedFunc(inputChan, cs.quitChan)

	loop:
		for {
			select {
			case <-cs.quitChan:
				break loop

			case res, ok := <-inputChan:
				if !ok {
					break loop
				}

				select {
				case <-cs.quitChan:
					break loop
				default:
				}

				if res.Err != nil {
					cs.errors = append(cs.errors, res.Err)
				}

				if !cs.hasError && res.Err != nil {
					cs.hasError = true
					cs.dataChannel <- res
					if cs.ape == stop {
						cs.Cancel()
					}
					continue
				}

				if cs.hasError && cs.ape == stop {
					continue
				}

				cs.dataChannel <- res
			}
		}

		safeCloseChannel(cs.dataChannel)

	}()

	return cs
}

// An option means stop the stream when has error
func StopWhenHasError() func(p *ChannelStream) {
	return func(p *ChannelStream) {
		p.ape = stop
	}
}

// An option means resume the stream when has error
func ResumeWhenHasError() func(p *ChannelStream) {
	return func(p *ChannelStream) {
		p.ape = resume
	}
}

// An option to set the count of go routines in the stream
func SetWorkers(workers int) func(p *ChannelStream) {
	return func(p *ChannelStream) {
		p.workers = workers
	}
}

func passByQuitChan(quitChan chan struct{}) func(p *ChannelStream) {
	return func(p *ChannelStream) {
		p.quitChan = quitChan
	}
}

// Pipe current steam output as another stream's input
func (p *ChannelStream) Pipe(dataPipeFunc PipeFunc, optionFuncs ...OptionFunc) *ChannelStream {
	seedFunc := func(dataPipeChannel chan<- Item, quitChannel chan struct{}) {
		wg := &sync.WaitGroup{}
		wg.Add(p.workers)
		for i := 0; i < p.workers; i++ {
			go func() {
				defer wg.Done()
			loop:
				for {
					select {
					case <-quitChannel:
						break loop
					case data, ok := <-p.dataChannel:
						if !ok {
							break loop
						}

						select {
						case <-quitChannel:
							break loop
						default:
						}

						dataPipeChannel <- dataPipeFunc(data)
					}
				}
			}()
		}

		go func() {
			wg.Wait()
			safeCloseChannel(dataPipeChannel)
		}()
	}

	mergeOptionFuncs := make([]OptionFunc, len(p.optionFuncs)+len(optionFuncs)+1)
	copy(mergeOptionFuncs[0:len(p.optionFuncs)], p.optionFuncs)
	copy(mergeOptionFuncs[len(p.optionFuncs):], optionFuncs)
	mergeOptionFuncs[len(p.optionFuncs)+len(optionFuncs)] = passByQuitChan(p.quitChan)

	return NewChannelStream(seedFunc, mergeOptionFuncs...)
}

func safeCloseChannel(dataPipeChannel chan<- Item) {
	if len(dataPipeChannel) == 0 {
		close(dataPipeChannel)
	} else {
		ticker := time.NewTicker(1 * time.Millisecond)
		for range ticker.C {
			if len(dataPipeChannel) == 0 {
				close(dataPipeChannel)
				break
			}
		}
		ticker.Stop()
	}
}

// Cancel current stream
func (p *ChannelStream) Cancel() {
	select {
	case _, ok := <-p.quitChan:
		if ok {
			close(p.quitChan)
		}
	default:
		close(p.quitChan)
	}

}

// Set race condition of current stream's output
func (p *ChannelStream) Race(raceFunc RaceFunc) {
loop:
	for item := range p.dataChannel {
		if raceFunc(item) {
			p.Cancel()
			break loop
		}
	}

	go func() {
		p.Drain()
	}()
}

// Harvest the output of current stream
func (p *ChannelStream) Harvest(harvestFunc HarvestFunc) (bool, []error) {
	for item := range p.dataChannel {
		harvestFunc(item)
	}

	return !p.hasError, p.errors
}

// Drain the output of current stream to make sure all the items got processed
func (p *ChannelStream) Drain() (bool, []error) {
	for range p.dataChannel {
	}

	return !p.hasError, p.errors
}
