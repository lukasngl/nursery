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

// This code block, "finsishes" only iff all jobs started with "StartSoon" are done.
result := nursery.WithUnbounded(func(nursery nursery.Executor[string]) {
	nursery.StartSoon(func() string {
		time.Sleep(2 * time.Millisecond)
		return "World"
	})

	nursery.StartSoon(func() string {
		return "Hello"
	})
})

fmt.Println(result)
// Output: [Hello World]
```
