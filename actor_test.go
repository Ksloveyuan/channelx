package channelx_test

import (
	"testing"
	"time"

	"github.com/Ksloveyuan/channelx"
)

func TestActorBasic(t *testing.T) {
	actor := channelx.NewActor(channelx.SetActorBuffer(1))
	defer actor.Close()

	promise := actor.Do(func() (interface{}, error) {
		time.Sleep(1 * time.Second)
		return 1, nil
	})

	res, err := promise.Done()

	if err != nil {
		t.Fail()
	}

	if res != 1 {
		t.Fail()
	}
}

func TestActorAsQueue(t *testing.T) {
	actor := channelx.NewActor()
	defer actor.Close()

	i := 0
	workFunc := func() (interface{}, error) {
		time.Sleep(1 * time.Second)
		i++
		return i, nil
	}

	promise := actor.Do(workFunc)
	promise2 := actor.Do(workFunc)

	res2, _ := promise2.Done()
	res1, _ := promise.Done()

	if res1 != 1 {
		t.Fail()
	}

	if res2 != 2 {
		t.Fail()
	}
}
