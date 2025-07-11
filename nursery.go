/*
Package nursery provides an opinionated abstraction over the common
fan-out-fan-in pattern inspired by [structured concurrency].

A nursery executes jobs given as closures and collects their results.
The nursery always waits until all jobs are completed.

For jobs with multiple return values a simple Tuple struct is provided,
removing the need for adhoc structs.

[structured concurrency]: https://vorpus.org/blog/notes-on-structured-concurrency-or-go-statement-considered-harmful/
*/
package nursery

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/sync/semaphore"
)

type Go[R any] = func(job func() R)

type Unbounded[R any] struct {
	mx              sync.Mutex
	done            bool
	resultC         chan R
	results         []R
	jobs            sync.WaitGroup
	resultCollector sync.WaitGroup
}

type Bounded[R any] struct {
	inner     *Unbounded[R]
	sem       *semaphore.Weighted
	scheduler sync.WaitGroup
	//nolint:containedctx // required for the semaphore
	ctx context.Context
}

// Tuple is an adapter type, to allow using functions with multiple returns types.
type Tuple[A, B any] struct {
	First  A
	Second B
}

// Unpack can be used to convinently assign tuple components to variables, e.g.
//
//	result, err := tuple.Unpack()
func (t Tuple[A, B]) Unpack() (A, B) {
	return t.First, t.Second
}

func NewTuple[A, B any](a A, b B) Tuple[A, B] {
	return Tuple[A, B]{a, b}
}

// WithBounded is the bounded variant of [WithUnbounded].
func WithBounded[R any](ctx context.Context, n int, run func(Go Go[R])) []R {
	nursery := NewBounded[R](ctx, n)

	run(nursery.Go)

	return nursery.Wait()
}

// WithUnbounded runs the code block given via the closure with a new nursery
// and waits for all started tasks to complete.
func WithUnbounded[R any](run func(Go Go[R])) []R {
	nursery := NewUnbounded[R]()

	run(nursery.Go)

	return nursery.Wait()
}

// NewUnbounded returns a new nursery, that executes at all jobs in parallel.
func NewUnbounded[R any]() *Unbounded[R] {
	nursery := &Unbounded[R]{
		resultC:         make(chan R),
		mx:              sync.Mutex{},
		done:            false,
		results:         []R{},
		jobs:            sync.WaitGroup{},
		resultCollector: sync.WaitGroup{},
	}

	nursery.resultCollector.Add(1)

	go func() {
		defer nursery.resultCollector.Done()

		for err := range nursery.resultC {
			nursery.results = append(nursery.results, err)
		}
	}()

	return nursery
}

// NewBounded returns a new nursery, that executes at most n jobs in parallel.
// Other jobs are scheduled and will wait until they are executed or the context is cancelled.
//
//nolint:varnamelen // n is perfectly fine
func NewBounded[R any](ctx context.Context, n int) *Bounded[R] {
	if n < 1 {
		panic(fmt.Sprintf("bound must be at least 1, but was %d", n))
	}

	return &Bounded[R]{
		ctx:       ctx,
		inner:     NewUnbounded[R](),
		sem:       semaphore.NewWeighted(int64(n)),
		scheduler: sync.WaitGroup{},
	}
}

// Go runs the code given via the closure in the background and collects its result.
func (nursery *Unbounded[R]) Go(job func() R) {
	nursery.startSoon(func() {
		nursery.resultC <- job()
	})
}

// Go runs the code given via the closure in the background and collects its result.
// If no more jobs can be run, because bounds are exceeded, the jobs gets scheduled and executed
// once other jobs finish.
// If the [Bounded] nursery's context is finished, the scheduled jobs will not be run.
func (nursery *Bounded[R]) Go(job func() R) {
	nursery.scheduler.Add(1)
	nursery.inner.startSoon(func() {
		defer nursery.scheduler.Done()

		err := nursery.sem.Acquire(nursery.ctx, 1)
		if err != nil {
			return
		}
		defer nursery.sem.Release(1)

		nursery.inner.resultC <- job()
	})
}

func (nursery *Unbounded[R]) startSoon(job func()) {
	nursery.mx.Lock()
	defer nursery.mx.Unlock()

	if nursery.done {
		panic("nursery is closed")
	}

	nursery.jobs.Add(1)

	go func() {
		defer nursery.jobs.Done()

		job()
	}()
}

// Wait blocks and returns all the collected results, once all jobs are finished.
func (nursery *Unbounded[R]) Wait() []R {
	nursery.mx.Lock()
	defer nursery.mx.Unlock()

	return nursery.wait()
}

// Wait blocks and returns all the collected results, once all jobs are finished.
func (nursery *Bounded[R]) Wait() []R {
	nursery.inner.mx.Lock()
	defer nursery.inner.mx.Unlock()

	// Wait until scheduled jobs are cleared
	nursery.scheduler.Wait()

	return nursery.inner.wait()
}

func (nursery *Unbounded[R]) wait() []R {
	if nursery.done {
		return nil
	}

	nursery.done = true

	nursery.jobs.Wait()

	close(nursery.resultC) // Note: closing the channel will stop the errCollector

	nursery.resultCollector.Wait()

	return nursery.results
}
