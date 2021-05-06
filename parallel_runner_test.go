package channelx

import (
	"errors"
	"fmt"
	"testing"
)

func TestRunInParallel(t *testing.T) {
	t.Run("normal case", func(t *testing.T) {
		worker := func(input interface{}) (interface{}, error) {
			num := input.(int)
			return num+1, nil
		}

		inputs := []interface{}{1,2,3,4,5}
		outputs, err := RunInParallel(inputs, worker, 4)

		if err != nil {
			t.Fail()
		}

		fmt.Println(outputs)
	})

	t.Run("if has error ", func(t *testing.T) {
		fakeErr := errors.New("fake error")
		worker := func(input interface{}) (interface{}, error) {
			num := input.(int)
			if num % 2 == 0 {
				return input, fakeErr
			}

			return num+1, nil
		}

		inputs := []interface{}{1,1,1,1,4}
		_, err := RunInParallel(inputs, worker, 4)

		if err != fakeErr {
			t.Fail()
		}
	})
}