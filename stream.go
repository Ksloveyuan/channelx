package channelx

import (
	"runtime"
	"sync"
	"time"
)

type ChannelStream struct {
	dataChannel chan Item
	workers     int
	ape         actionPerError
	optionFuncs []OptionFunc
	hasError    bool
	errors      []error
	quitChan    chan struct{}
}

type Item struct {
	Data interface{}
	Err  error
}

type actionPerError int

const (
	resume actionPerError = 0
	stop                  = 1
)

type SeedFunc func(seedChan chan<- Item, quitChannel chan struct{})
type PipeFunc func(item Item) Item
type HarvestFunc func(item Item)
type RaceFunc func(item Item) bool
type OptionFunc func(cs *ChannelStream)

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

func StopWhenHasError() func(p *ChannelStream) {
	return func(p *ChannelStream) {
		p.ape = stop
	}
}

func ResumeWhenHasError() func(p *ChannelStream) {
	return func(p *ChannelStream) {
		p.ape = resume
	}
}

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

func (p *ChannelStream) Harvest(harvestFunc HarvestFunc) (bool, []error) {
	for item := range p.dataChannel {
		harvestFunc(item)
	}

	return !p.hasError, p.errors
}

func (p *ChannelStream) Drain() (bool, []error) {
	for range p.dataChannel {
	}

	return !p.hasError, p.errors
}
