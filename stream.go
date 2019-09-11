package channelutil

import (
	"context"
	"runtime"
	"sync"
	"time"
)

type ChannelStream struct {
	dataChannel chan Result
	workers     int
	ape         actionPerError
	optionFuncs []OptionFunc
	hasError    bool
	errors      []error
}

type Result struct {
	Data interface{}
	Err  error
}

type actionPerError int

const (
	resume actionPerError = 0
	stop                  = 1
)

type SeedFunc func(seedChan chan<- Result)
type PipeFunc func(result Result) Result
type HarvestFunc func(result Result)
type RaceFunc func(result Result) bool

type OptionFunc func(cs *ChannelStream)

func NewChannelStream(ctx context.Context, seedFunc SeedFunc, optionFuncs ...OptionFunc) *ChannelStream {
	cs := &ChannelStream{
		workers:     runtime.NumCPU(),
		optionFuncs: optionFuncs,
	}

	for _, of := range optionFuncs {
		of(cs)
	}

	cs.dataChannel = make(chan Result, cs.workers)

	cancelCtx, cancel := context.WithCancel(ctx)

	go func() {
		inputChan := make(chan Result, cs.workers)

		go seedFunc(inputChan)

	loop:
		for {
			select {
			case <-cancelCtx.Done():
				select {
				case res := <-inputChan:
					if res.Err != nil {
						cs.errors = append(cs.errors, res.Err)
					}

					if !cs.hasError && res.Err != nil {
						cs.hasError = true
						cs.dataChannel <- res
						continue
					}

					if res.Err == nil || cs.ape == resume {
						cs.dataChannel <- res
					}
				default:
					break loop
				}
			case res, ok := <-inputChan:
				if !ok {
					break loop
				}

				if res.Err != nil {
					cs.errors = append(cs.errors, res.Err)
				}

				if !cs.hasError && res.Err != nil {
					cs.hasError = true
					cs.dataChannel <- res
					continue
				}

				if cs.hasError && cs.ape == stop {
					cancel()
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

func (p *ChannelStream) Pipe(ctx context.Context, dataPipeFunc PipeFunc, optionFuncs ...OptionFunc) *ChannelStream {
	cancelCtx, _ := context.WithCancel(ctx)
	seedFunc := func(dataPipeChannel chan<- Result) {
		wg := &sync.WaitGroup{}
		wg.Add(p.workers)
		for i := 0; i < p.workers; i++ {
			go func() {
				defer wg.Done()
			loop:
				for {
					select {
					case <-cancelCtx.Done():
						select {
						case data := <-p.dataChannel:
							dataPipeChannel <- dataPipeFunc(data)
						default:
							break loop
						}
					case data, ok := <-p.dataChannel:
						if !ok {
							break loop
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

	mergeOptionFuncs := make([]OptionFunc, len(p.optionFuncs)+len(optionFuncs))
	copy(mergeOptionFuncs[0:len(p.optionFuncs)], p.optionFuncs)
	copy(mergeOptionFuncs[len(p.optionFuncs):], optionFuncs)

	return NewChannelStream(cancelCtx, seedFunc, mergeOptionFuncs...)
}

func safeCloseChannel(dataPipeChannel chan<- Result) {
	if len(dataPipeChannel) == 0 {
		close(dataPipeChannel)
	} else {
		ticker := time.Tick(1 * time.Millisecond)
		for range ticker {
			if len(dataPipeChannel) == 0 {
				close(dataPipeChannel)
				break
			}
		}
	}
}

func (p *ChannelStream) Race(ctx context.Context, raceFunc RaceFunc) {
	_, cancel := context.WithCancel(ctx)

	for result := range p.dataChannel {
		if raceFunc(result) {
			cancel()
			break
		}
	}

	go func() {
		for range p.dataChannel {
		}
	}()
}

func (p *ChannelStream) Harvest(harvestFunc HarvestFunc) (bool, []error) {
	for result := range p.dataChannel {
		harvestFunc(result)
	}

	return !p.hasError, p.errors
}

func (p *ChannelStream) Drain() (bool, []error) {
	for range p.dataChannel {
	}

	return !p.hasError, p.errors
}
