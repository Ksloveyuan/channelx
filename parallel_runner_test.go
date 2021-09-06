package channelx_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/Ksloveyuan/channelx"
)

func TestRunInParallel(t *testing.T) {
	t.Run("normal case", func(t *testing.T) {
		worker := func(ctx context.Context, input interface{}) (interface{}, error) {
			num := input.(int)
			return num+1, nil
		}

		inputs := []interface{}{1,2,3,4,5}
		outputs, err := channelx.RunInParallel(context.Background(), inputs, worker, 4)

		if err != nil {
			t.Fail()
		}

		fmt.Println(outputs)
	})

	t.Run("if has error ", func(t *testing.T) {
		fakeErr := errors.New("fake error")
		worker := func(ctx context.Context, input interface{}) (interface{}, error) {
			num := input.(int)
			if num % 2 == 0 {
				return input, fakeErr
			}

			return num+1, nil
		}

		inputs := []interface{}{1,1,1,1,4}
		ctx := context.Background()
		_, err := channelx.RunInParallel(ctx, inputs, worker, 4)

		if err != fakeErr {
			t.Fail()
		}
	})
}