package channelutil

import (
	"sync"
)

type ChannelStream struct {
	dataChannel chan Result
	workers int
	ape actionPerError
	optionFuncs []OptionFunc
	hasError bool
}

type Result struct {
	Data interface{}
	Err error
}

type actionPerError int

const (
	resume actionPerError = 0
	stop = 1
)

type SeedFunc func(seedChan chan <- Result)
type PipeFunc func(result Result) Result
type HarvestFunc func(result Result)

type OptionFunc func(cs *ChannelStream) func()

func NewChannelStream(seedFunc SeedFunc, optionFuncs ...OptionFunc) *ChannelStream {
	cs := &ChannelStream{
		dataChannel: make(chan Result, 4),
		workers:     4,
		optionFuncs: optionFuncs,
	}

	for _, of := range optionFuncs {
		of(cs)()
	}

	go func() {
		inputChan := make(chan Result, 4)

		go seedFunc(inputChan)

		for res := range inputChan {
			if !cs.hasError && res.Err != nil{
				cs.hasError = true
			}

			if cs.hasError && cs.ape == stop {
				continue
			}

			cs.dataChannel <- res
		}
		close(cs.dataChannel)
	}()

	return cs
}

func StopWhenHasError(p *ChannelStream) func() {
	return func() {
		p.ape = stop
	}
}

func ResumeWhenHasError(p *ChannelStream) func() {
	return func() {
		p.ape = resume
	}
}

func (p *ChannelStream) Pipe(dataPipeFunc PipeFunc) *ChannelStream {
	seedFunc := func(dataPipeChannel chan <- Result) {
		wg := &sync.WaitGroup{}
		wg.Add(p.workers)
		for i:=0; i < p.workers; i++{
			go func() {
				defer wg.Done()
				for data := range p.dataChannel {
					dataPipeChannel <- dataPipeFunc(data)
				}
			}()
		}

		go func() {
			wg.Wait()
			close(dataPipeChannel)
		}()
	}

	return NewChannelStream(seedFunc, p.optionFuncs...)
}

func (p *ChannelStream) Done(harvestFunc HarvestFunc) {
	for result := range p.dataChannel {
		harvestFunc(result)
	}
}
