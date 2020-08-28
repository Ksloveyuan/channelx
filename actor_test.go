package channelx

import (
	"fmt"
	"testing"
	"time"
)

func TestActor(t *testing.T) {
	actor := NewActiveObject()
	defer actor.Close()

	actor.Call(func() (interface{}, error) {
		time.Sleep(1 * time.Second)
		return 1, nil
	})

	actor.CallWithCallback(func() (interface{}, error) {
		time.Sleep(1 * time.Second)
		return 1, nil
	}, func(i interface{}, err error) {
		if err != nil {
			fmt.Print(err)
		}
	})

	future := actor.CallWithFuture(func() (interface{}, error) {
		time.Sleep(1 * time.Second)
		return 1, nil
	})

	res, err := future.Done()

	if err != nil {
		t.Fail()
	}

	if res != 1 {
		t.Fail()
	}
}
