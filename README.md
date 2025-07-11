[![Go Reference](https://pkg.go.dev/badge/github.com/lukasngl/opt.svg)](https://pkg.go.dev/github.com/lukasngl/nursery)

Opinionated abstraction over the common fan-out fan-in pattern
inspired by [structured concurrency].

A nursery executes jobs given as closures and collects their results.
The nursery always waits until all jobs are completed.

[structured concurrency]: https://vorpus.org/blog/notes-on-structured-concurrency-or-go-statement-considered-harmful/

# Example

The packages export the `With*` functions, to intuitively run dynamic jobs
in the given closure, and wait until all jobs are done.

```go

// This code block, "finishes" only iff all jobs started with "Go" are done.
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
```
