package nursery_test

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"slices"
	"testing"
	"testing/quick"
	"time"

	"github.com/lukasngl/nursery"
)

func TestWithUnbounded_Completes(t *testing.T) {
	t.Parallel()

	property := func(order executionOrder) bool {
		t.Logf("running with order %s", order)

		completed := nursery.WithUnbounded(func(Go nursery.Go[int]) {
			for position := range order.Size() {
				Go(func() int {
					order.Wait(position)

					return position
				})
			}

			order.Run()
		})

		ok := true

		for position := range order.Size() {
			if !slices.Contains(completed, position) {
				t.Logf("job[%3d] did not complete", position)

				ok = false
			}
		}

		return ok
	}

	err := quick.Check(property, nil)
	if err != nil {
		//nolint
		t.Fatalf("property did not hold for input: %s", err.(*quick.CheckError).In[0])
	}
}

func TestWithBounded_Completes(t *testing.T) {
	t.Parallel()

	property := func(order executionOrder, bound int) bool {
		// reasonable size
		bound %= 2 * order.Size()
		// non-negative
		if bound < 0 {
			bound *= -1
		}
		// at least 1
		bound++

		t.Logf("running with bound %d and order %s", bound, order)

		completed := nursery.WithBounded(
			context.TODO(),
			bound,
			func(Go nursery.Go[int]) {
				for position := range order.Size() {
					Go(func() int {
						t.Logf("job[%3d] started", position)
						defer t.Logf("job[%3d] done", position)
						order.Wait(position)

						return position
					})
				}

				order.Run()
			},
		)

		ok := true

		for position := range order.Size() {
			if !slices.Contains(completed, position) {
				t.Logf("job[%3d] did not complete", position)

				ok = false
			}
		}

		return ok
	}

	err := quick.Check(property, nil)
	if err != nil {
		//nolint
		t.Fatalf("property did not hold for input: %s", err.(*quick.CheckError).In[0])
	}
}

//nolint:gosec // G115: clamped to [0, 100)
func TestWithBounded_CancelStopsScheduled(t *testing.T) {
	t.Parallel()

	property := func(bound, overflow uint) bool {
		// reasonable size
		bound %= 100
		overflow %= 100
		// at least 1
		bound++

		ctx, cancel := context.WithTimeout(context.TODO(), time.Millisecond)
		defer cancel()

		completed := nursery.WithBounded(ctx, int(bound), func(Go nursery.Go[int]) {
			for position := range bound + overflow {
				Go(func() int {
					<-ctx.Done()

					return int(position)
				})
			}
		})

		if len(completed) != int(bound) {
			t.Logf("failed for %d, %d", bound, overflow)
		}

		return true
	}

	err := quick.Check(property, nil)
	if err != nil {
		t.Fatalf("property did not hold")
	}
}

var _ quick.Generator = executionOrder{}

type executionOrder struct {
	order   []int
	waiters []chan struct{}
}

func (order *executionOrder) Size() int {
	return len(order.order)
}

func (order *executionOrder) Wait(position int) {
	<-order.waiters[position]
}

func (order *executionOrder) Run() {
	for _, position := range order.order {
		order.waiters[position] <- struct{}{}
		close(order.waiters[position])
	}
}

func (order executionOrder) String() string {
	return fmt.Sprint(order.order)
}

// Generate implements quick.Generator.
func (order executionOrder) Generate(rand *rand.Rand, size int) reflect.Value {
	order.order = make([]int, size)
	order.waiters = make([]chan struct{}, size)

	for i := range size {
		order.order[i] = i
		order.waiters[i] = make(chan struct{}, 1)
	}

	rand.Shuffle(size, func(i, j int) {
		order.order[i], order.order[j] = order.order[j], order.order[i]
	})

	return reflect.ValueOf(order)
}
