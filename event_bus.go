package channelx

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"
)


//EventBus ...
type EventBus struct {
	started  bool
	mu       sync.RWMutex
	quit     chan struct{}
	wg       *sync.WaitGroup
	handlers map[EventID][]EventHandler

	logger         Logger
	hostName       string
	eventWorkers   int
	eventJobQueue  chan EventJob
	autoRetryTimes int
	retryInterval  time.Duration

	runningTasks *int64
}

// EventJob ...
type EventJob struct {
	event      Event
	handlers   []EventHandler
	resultChan chan JobStatus
}


//NewEventBus ...
func NewEventBus(
	logger Logger,
	hostName string,
	chanBuffer,
	eventWorkers,  autoRetryTimes int,
	retryInterval time.Duration,
) *EventBus {
	runningTasks := int64(0)
	return &EventBus{
		logger:         logger,
		hostName:       hostName,
		eventWorkers:   eventWorkers,
		autoRetryTimes: autoRetryTimes,
		retryInterval:  retryInterval,

		handlers:      make(map[EventID][]EventHandler),
		eventJobQueue: make(chan EventJob, chanBuffer),
		quit:          make(chan struct{}),
		wg:            &sync.WaitGroup{},

		runningTasks: &runningTasks,
	}
}

// Start ...
func (eb *EventBus) Start() {
	if eb.started {
		return
	}

	for i := 0; i < eb.eventWorkers; i++ {
		eb.wg.Add(1)
		go eb.eventWorker(eb.eventJobQueue)
	}

	eb.started = true
}

// Stop ....
func (eb *EventBus) Stop() {
	if !eb.started {
		return
	}

	close(eb.quit)
	eb.wg.Wait()
	eb.started = false
}

func (eb *EventBus) eventWorker(jobQueue <-chan EventJob) {
	defer func() {
		if r := recover(); r != nil {
			eb.logger.Errorf("eventWorker recovered, error is %v", r)
			atomic.AddInt64(eb.runningTasks, -1)
			eb.eventWorker(jobQueue)
		}
	}()

loop:
	for {
		select {
		case job := <-jobQueue:
			atomic.AddInt64(eb.runningTasks, 1)
			jobStatus := JobStatus{
				RunAt: time.Now(),
			}
			eb.logger.Debugf("event start to run, eventID is %d", job.event.ID())

			if len(job.handlers) > 1 {
				g, _ := errgroup.WithContext(context.Background())
				for index := range job.handlers {
					handler := job.handlers[index]
					g.Go(func() error {
						return eb.runHandler(handler, job.event)
					})
				}
				jobStatus.Err = g.Wait()
			} else if len(job.handlers) == 1 {
				jobStatus.Err = eb.runHandler(job.handlers[0], job.event)
			}

			if jobStatus.Err != nil {
				eb.logger.Errorf("event failed, eventID is %d, error is %v", job.event.ID(), jobStatus.Err)
			} else {
				eb.logger.Infof("event succeed, eventID is %d", job.event.ID())
			}

			jobStatus.FinishedAt = time.Now()

			eb.logger.Debugf("event finished, eventID is %d", job.event.ID())

			select {
			case job.resultChan <- jobStatus:
			default:
				eb.logger.Warnf("job's result channel is blocked, eventID is %d", job.event.ID())
			}

			eb.logger.Debugf("event result is sent, eventID is %d", job.event.ID())
			atomic.AddInt64(eb.runningTasks, -1)
		case <-eb.quit:
			if eb.anyPendingTask() {
				continue loop
			}

			eb.wg.Done()
			eb.logger.Debugf("event worker quit")
			break loop
		}
	}
}

func (eb *EventBus) runHandler(handler EventHandler, event Event) (err error) {
	var (
		retryDuration = eb.retryInterval
	)

	defer func() {
		if r := recover(); r != nil {
			eb.logger.Errorf("event handler recovered, err is %v", r)
			err = fmt.Errorf("%v", r)
		}
	}()

	for i := 0; i <= eb.autoRetryTimes; i++ {
		if err != nil {
			eb.logger.Infof("start to retry event handler %T with event %T, times(%d), last error(%v)", handler, event, i, err)
		}

		if err = handler.OnEvent(event); err == nil {
			break
		}

		eb.logger.Errorf("error on event handler %T with event %T: %v", handler, event, err)
		if !handler.CanAutoRetry(err) {
			break
		}

		time.Sleep(retryDuration)
		retryDuration *= 2
	}
	return err
}

func (eb *EventBus) anyPendingTask() bool {
	return atomic.LoadInt64(eb.runningTasks) != 0 ||
		len(eb.eventJobQueue) > 0
}

//Subscribe ...
func (eb *EventBus) Subscribe(eventID EventID, handlers ...EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.handlers[eventID] = append(eb.handlers[eventID], handlers...)
}

//Unsubscribe ...
func (eb *EventBus) Unsubscribe(eventID EventID, handlers ...EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if _, ok := eb.handlers[eventID]; ok {
		for i, h := range eb.handlers[eventID] {
			for _, H := range handlers {
				if h == H {
					eb.handlers[eventID] = append(eb.handlers[eventID][:i], eb.handlers[eventID][i+1:]...)
				}
			}
		}
	}
}

//Publish ...
func (eb *EventBus) Publish(evt Event) <-chan JobStatus {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	if ehs, ok := eb.handlers[evt.ID()]; ok {
		handlers := make([]EventHandler, len(ehs))
		copy(handlers, ehs)
		job := EventJob{
			event:      evt,
			handlers:   handlers,
			resultChan: make(chan JobStatus, 1),
		}

		var jobQueue = eb.eventJobQueue
		select {
		case jobQueue <- job:
		default:
			go func() {
				eb.logger.Warnf("job queue is full, waiting to enqueue")
				atomic.AddInt64(eb.runningTasks, 1)
				jobQueue <- job
				atomic.AddInt64(eb.runningTasks, -1)
				eb.logger.Infof("finally enqueue successfully")
			}()
		}

		return job.resultChan
	} else {
		err := fmt.Errorf("no handlers for event(%d)", evt.ID())
		resultChan := make(chan JobStatus, 1)
		resultChan <- JobStatus{
			Err: err,
		}
		eb.logger.Errorf("no handlers for event(%d)", evt.ID())
		return resultChan
	}
}
