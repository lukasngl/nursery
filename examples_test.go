package nursery_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/lukasngl/nursery"
)

func someOperation() nursery.Tuple[string, error] {
	return nursery.NewTuple[string, error]("success", nil)
}

func ExampleTuple_Unpack() {
	result, err := someOperation().Unpack()
	if err != nil {
		panic(err)
	}

	fmt.Println(result)
	// Output: success
}

func ExampleWithUnbounded() {
	result := nursery.WithUnbounded(func(Go nursery.Go[string]) {
		Go(func() string {
			time.Sleep(2 * time.Millisecond)
			return "World"
		})

		Go(func() string {
			return "Hello"
		})
	})

	fmt.Println(result)
	// Output: [Hello World]
}

func operationThatWillTimout(ctx context.Context) (string, error) {
	<-ctx.Done()

	//nolint:err113 // just an example
	return "", errors.New("timed out :/")
}

const someTimeout = time.Millisecond

func ExampleWithBounded() {
	ctx, cancel := context.WithTimeout(context.TODO(), someTimeout)
	defer cancel()

	const maxParallel = 2

	results := nursery.WithBounded(
		ctx,
		maxParallel,
		func(Go nursery.Go[nursery.Tuple[string, error]]) {
			for range 10 {
				Go(func() nursery.Tuple[string, error] {
					return nursery.NewTuple(operationThatWillTimout(ctx))
				})
			}
		},
	)

	fmt.Println(len(results))

	for i := range results {
		result, err := results[i].Unpack()
		fmt.Printf("result=%s err=%s\n", result, err)
	}

	// Output: 2
	// result= err=timed out :/
	// result= err=timed out :/
}
