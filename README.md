# SIS - Text Analyzer (Checkpoint 1)

This is my SIS project for the Go course. It is a small in-memory app where a "task" means analyzing a piece of text. You give it a string and it counts the words, the characters and finds the longest word.

For example, for `"Hello Go world"` the result is 3 words, 14 characters and the longest word is `Hello`.

Everything is stored in memory and runs synchronously, there is no database and no concurrency yet.

## How the code is organized

- `main.go` - a short demo that creates a task, processes it and prints the result
- `task/task.go` - the `Task` struct, the statuses, the errors and the rules for status changes
- `task/processor.go` - the `Processor` interface and the `TextAnalyzer` that does the actual counting
- `task/store.go` - the `Store` that keeps tasks in a map and has `Create`, `Get` and `Process`
- `task/task_test.go` - unit tests

## The Task

A task has an `ID` (an int that the store generates), the input `Text`, a `Status` and a `Result`. The result is a pointer, so it is `nil` until the task has been processed.

## Task lifecycle

A task starts as `Pending`, then goes to `Processing` and ends as `Completed`. If the processor returns an error, it ends as `Failed` instead.

```
Pending -> Processing -> Completed
                     \-> Failed
```

Allowed transitions:

- `Pending` can only go to `Processing`
- `Processing` can go to `Completed` or `Failed`
- `Completed` and `Failed` are final, nothing can follow them

Anything else returns `ErrInvalidTransition`. For example, if you try to process a task that is already completed, you get this error and the task stays as it was.

## Validation

The text must not be empty or only whitespace. If it is, `Create` returns `ErrEmptyInput` and no task is created.

## Errors

I did not use `panic` anywhere. Expected problems are returned as normal errors:

- `ErrEmptyInput` for bad input
- `ErrTaskNotFound` when there is no task with that ID
- `ErrInvalidTransition` for a wrong status change

I wrap them with `%w` so I can add details like the task ID and still check the type with `errors.Is`.

## Some design choices

I made `Processor` an interface so the `Store` does not depend on one specific analyzer. This also helped in the tests, because I could pass in a fake processor that always fails and check that the task becomes `Failed`.

Tasks are kept in a `map[int]*Task`, so finding a task by ID is fast. I store pointers so that when I change the status or the result, it changes the real task in the map and not a copy.

`Task.transition` has a pointer receiver because it changes the task. `TextAnalyzer.Process` has a value receiver because that struct has no fields to change.

For counting characters I used `utf8.RuneCountInString` instead of `len`, because `len` counts bytes and would give wrong numbers for non-English text like Russian.

## Running and checking

```
go run .
gofmt -l .
go vet ./...
go test ./...
```

## Tests

The tests use the standard `testing` package. They check:

- creating a task
- invalid input (table-driven)
- getting a task, and getting one that does not exist
- processing, the result values and the final status (table-driven)
- processing a completed task again, which must fail
- a processor error, which must give the `Failed` status
- allowed and not allowed status transitions (table-driven)
- several tasks at once, to make sure their data does not get mixed up
