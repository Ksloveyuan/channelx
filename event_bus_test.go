package channelx_test

import (
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
		eventBus := channelx.NewEventBus(logger, "test", 4,4,2, time.Second)
		eventBus.Start()

		fakeEvent := channelx_mock.NewMockEvent(ctl)
		fakeEvent.EXPECT().ID().Return(fakeEventID).AnyTimes()

		waitGroup := sync.WaitGroup{}
		fakeHandler := channelx_mock.NewMockEventHandler(ctl)
		fakeHandler.EXPECT().OnEvent(fakeEvent).Do(func(_ interface{}) {
			waitGroup.Done()
		})

		eventBus.Subscribe(fakeEventID, fakeHandler)
		waitGroup.Add(1)
		eventBus.Publish(fakeEvent)

		waitGroup.Wait()
		eventBus.Stop()
	})

	t.Run("auto retry", func(t *testing.T) {
		eventBus := channelx.NewEventBus(logger, "test", 4,4,2, time.Second)
		eventBus.Start()

		fakeError := fmt.Errorf("fake error")
		fakeEvent := channelx_mock.NewMockEvent(ctl)
		fakeEvent.EXPECT().ID().Return(fakeEventID).AnyTimes()

		waitGroup := sync.WaitGroup{}
		fakeHandler := channelx_mock.NewMockEventHandler(ctl)
		fakeHandler.EXPECT().CanAutoRetry(fakeError).Return(true).Times(3)
		fakeHandler.EXPECT().OnEvent(fakeEvent).DoAndReturn(func(_ interface{}) error {
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
