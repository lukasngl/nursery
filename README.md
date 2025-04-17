Opinionated abstraction over the common fan-out fan-in pattern inspired by [structured concurrency].

A nursery executes jobs given as closures and collects their results.
The nursery always waits until all jobs are completed.

[structured concurrency]: https://vorpus.org/blog/notes-on-structured-concurrency-or-go-statement-considered-harmful/

# Example

```go
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
