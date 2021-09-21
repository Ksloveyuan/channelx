package channelx_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Ksloveyuan/channelx"
	channelx_mock "github.com/Ksloveyuan/channelx/mocks"
	"github.com/golang/mock/gomock"
)

func TestEventBus(t *testing.T) {
	const (
		fakeEventID channelx.EventID = 1
	)
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	logger := channelx.NewConsoleLogger()

	t.Run("happy flow", func(t *testing.T) {
		eventBus := channelx.NewEventBus(logger,  4,4,2, time.Second, 5*time.Second)
		eventBus.Start()

		fakeEvent := channelx_mock.NewMockEvent(ctl)
		fakeEvent.EXPECT().ID().Return(fakeEventID).AnyTimes()

		waitGroup := sync.WaitGroup{}
		fakeHandler := channelx_mock.NewMockEventHandler(ctl)
		fakeHandler.EXPECT().OnEvent(gomock.Any(), fakeEvent).Do(func(_,_ interface{}) {
			waitGroup.Done()
		})

		eventBus.Subscribe(fakeEventID, fakeHandler)
		waitGroup.Add(1)
		eventBus.Publish(fakeEvent)

		waitGroup.Wait()
		eventBus.Stop()
	})

	t.Run("timout", func(t *testing.T) {
		eventBus := channelx.NewEventBus(logger,  4,4,0, time.Second, time.Second)
		eventBus.Start()

		fakeError := fmt.Errorf("timeout")
		fakeEvent := channelx_mock.NewMockEvent(ctl)
		fakeEvent.EXPECT().ID().Return(fakeEventID).AnyTimes()

		waitGroup := sync.WaitGroup{}
		fakeHandler := channelx_mock.NewMockEventHandler(ctl)
		fakeHandler.EXPECT().CanAutoRetry(fakeError).Return(true)
		fakeHandler.EXPECT().OnEvent(gomock.Any(), fakeEvent).DoAndReturn(func(ctx context.Context, event channelx.Event) error {
			select {
			case <-ctx.Done():
				waitGroup.Done()
				logger.Infof("timeout")
				return fakeError
			case <-time.After(2 * time.Second):
				return nil
			}
		})

		eventBus.Subscribe(fakeEventID, fakeHandler)
		waitGroup.Add(1)
		eventBus.Publish(fakeEvent)

		waitGroup.Wait()
		eventBus.Stop()
	})

	t.Run("auto retry", func(t *testing.T) {
		eventBus := channelx.NewEventBus(logger,  4,4,2, time.Second, 5 * time.Second)
		eventBus.Start()

		fakeError := fmt.Errorf("fake error")
		fakeEvent := channelx_mock.NewMockEvent(ctl)
		fakeEvent.EXPECT().ID().Return(fakeEventID).AnyTimes()

		waitGroup := sync.WaitGroup{}
		fakeHandler := channelx_mock.NewMockEventHandler(ctl)
		fakeHandler.EXPECT().CanAutoRetry(fakeError).Return(true).Times(3)
		fakeHandler.EXPECT().OnEvent(gomock.Any(), fakeEvent).DoAndReturn(func(_,_ interface{}) error {
			waitGroup.Done()
			return fakeError
		}).Times(3)

		eventBus.Subscribe(fakeEventID, fakeHandler)
		waitGroup.Add(3)
		resultChan := eventBus.Publish(fakeEvent)

		waitGroup.Wait()
		result :=  <-resultChan
		if result.Err != fakeError {
			t.Fail()
		}
		eventBus.Stop()
	})
}

func TestEventBus_Example(t *testing.T) {
	logger := channelx.NewConsoleLogger()
	eventBus := channelx.NewEventBus(logger,  4,4,2, time.Second, 5 * time.Second)
	eventBus.Start()

	handler := NewExampleHandler(logger)
	eventBus.Subscribe(ExampleEventID, handler)
	eventBus.Publish(NewExampleEvent())

	eventBus.Stop()
}

const ExampleEventID channelx.EventID = 1

type ExampleEvent struct {
	id channelx.EventID
}

func NewExampleEvent() ExampleEvent {
	return ExampleEvent{id:ExampleEventID}
}

func (evt ExampleEvent) ID() channelx.EventID  {
	return evt.id
}

type ExampleHandler struct {
	logger channelx.Logger
}

func NewExampleHandler(logger channelx.Logger) *ExampleHandler {
	return &ExampleHandler{
		logger: logger,
	}
}

func (h ExampleHandler) Logger() channelx.Logger{
	return h.logger
}

func (h ExampleHandler) CanAutoRetry(err error) bool {
	return false
}

func (h ExampleHandler) OnEvent(ctx context.Context, event channelx.Event) error {
	if event.ID() != ExampleEventID {
		return fmt.Errorf("subscribe wrong event(%d)", event.ID())
	}

	_, ok := event.(ExampleEvent)
	if !ok {
		return fmt.Errorf("failed to convert received event to ExampleEvent")
	}

	h.Logger().Infof("event handled")

	return nil
}
