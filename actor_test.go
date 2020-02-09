package channelx

import (
	"testing"
	"time"
)

func TestActorBasic(t *testing.T) {
	actor := NewActor(SetActorBuffer(1))
	defer actor.Close()

	call := actor.Do(func() (interface{}, error) {
		time.Sleep(1 * time.Second)
		return 1, nil
	})

	res, err := call.Done()

	if err != nil {
		t.Fail()
	}

	if res != 1 {
		t.Fail()
	}
}

func TestActorAsQueue(t *testing.T) {
	actor := NewActor()
	defer actor.Close()

	i := 0
	workFunc := func() (interface{}, error) {
		time.Sleep(1 * time.Second)
		i++
		return i, nil
	}

	call1 := actor.Do(workFunc)
	call2 := actor.Do(workFunc)

	res2, _ := call2.Done()
	res1, _ := call1.Done()

	if res1 != 1 {
		t.Fail()
	}

	if res2 != 2 {
		t.Fail()
	}
}
