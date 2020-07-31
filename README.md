# taskq — a lightweight asynchronous task queue
[![GoDoc](https://godoc.org/github.com/airnandez/taskq?status.svg)](https://godoc.org/github.com/airnandez/taskq)

## Overview
`taskq` is a intentionally thin layer on top of the standard library for enqueueing and executing tasks asynchronously and concurrently.
Internally, a `taskq` owns a set of worker goroutines which each executes one task at a time.

## How to use
As a user, when you create a new `taskq` object, you can specify the number of worker gouroutines and the size of the waiting area of the queue.

```go
import (
    "fmt"

    "github.com/airnandez/taskq"
)

// A 'task' must implement the taskq.Task interface which is defined as
//
//    type Task interface {
//       Execute()
//    }
//
type SimpleTask int

func (t SimpleTask) Execute() {
    fmt.Printf("Hello I'm task %d\n", t)
}

func ExampleSimple() {
    // Create a queue with 2 workers
    q := taskq.New(taskq.Workers(2))

    // Submit 5 tasks for execution.
    for i := 0; i < 5; i++ {
        task := SimpleTask(i)
        q.Submit(task)
    }

    // Notify the queue that no more tasks will be submitted
    // and wait for all the already submitted tasks to finish
    q.CloseAndWait()

    // Unordered output:
    // Hello I'm task 0
    // Hello I'm task 1
    // Hello I'm task 2
    // Hello I'm task 3
    // Hello I'm task 4
}
```

In the documentation you can find a complete [executable example](https://godoc.org/github.com/airnandez/taskq#pkg-examples).

## Installation
You need to have installed the [Go programming environment](https://golang.org) and then do:

```
go get -u github.com/airnandez/taskq
```

## Status

Experimental.

## Feedback

Constructive feedback is welcome. Please feel free to provide it by [opening an issue](https://github.com/airnandez/taskq/issues).

## Credits

This tool is being developed and maintained by Fabio Hernandez at [IN2P3 / CNRS computing center](http://cc.in2p3.fr) (Lyon, France).

## License
Copyright 2020 Fabio Hernandez

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
