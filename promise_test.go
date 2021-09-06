package channelx_test

import (
	"errors"
	"testing"
	"time"

	"github.com/Ksloveyuan/channelx"
)

func TestPromise(t *testing.T) {
	promise := channelx.NewPromise(func() (interface{}, error) {
		time.Sleep(1 * time.Second)
		return 1, nil
	})

	res, _ := promise.Done()
	if res != 1 {
		t.Fail()
	}
}

func TestPromise_ThenSuccess(t *testing.T) {
	promise := channelx.NewPromise(func() (interface{}, error) {
		time.Sleep(1 * time.Second)
		return 1, nil
	}).ThenSuccess(func(input interface{}) (interface{}, error) {
		i := input.(int)
		return i + 1, nil
	})

	res, _ := promise.Done()
	if res != 2 {
		t.Fail()
	}
}

func TestPromise_ThenError(t *testing.T) {
	promise := channelx.NewPromise(func() (interface{}, error) {
		time.Sleep(1 * time.Second)
		return nil, errors.New("test")
	}).ThenError(func(err error) interface{} {
		if err == nil {
			t.Fail()
		}
		return 1
	})

	res, _ := promise.Done()
	if res != 1 {
		t.Fail()
	}
}

func TestPromise_Then(t *testing.T) {
	promise := channelx.NewPromise(func() (interface{}, error) {
		time.Sleep(1 * time.Second)
		return 1, nil
	}).Then(func(input interface{}) (interface{}, error) {
		return nil, errors.New("test")
	}, func(err error) interface{} {
		if err == nil {
			t.Fail()
		}
		return 1
	})

	_, err := promise.Done()
	if err == nil {
		t.Fail()
	}
}
