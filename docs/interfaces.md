# Golang SDK

The current implementation provides a Golang SDK. In order to be able to use the simulator, the user can implement the **Node** interface.

Each interface comes with a **Util** interface provided to the user for interacting with a simulation:
```go
// This interface needs to be implemented by a node.
type Node interface {
	// constructor interface
	New(util NodeUtil) Node

	// Function that represents what the node should do when notified.
	OnNotify()

	// Function that represent the initial action taken by a node when it
	// joins the network.
	OnJoin()

	// A method that should be called when a node leaves the network.
	OnLeave()
}
```

A user needs to implement all the methods provided below to obtain a valid node to use in a simulation:
```go
import (
  . "github.com/srene/Speer/interfaces"
)

type Example struct {
  //[...]
}

func (s *Example) OnJoin() {
  //[...]
}

func (s *Example) OnNotify() {
  //[...]
}

func (s *Example) OnLeave() {
  //[...]
}

func (s *Example) New(util Util) Node {
  //[...]
}
```

The `New` function is used to generate new nodes from an empty structure template. The structure `NodeUtil` provides a set of functions which `Example` can use to interact with the simulation:
```go
// The Util interface is provided to a node.
type NodeUtil interface {
	RoutineCapabilities

	Transport() Transport // Interface used to send data to other nodes.

	Id() string // The ID of the node.
	Join() string  // The ID of another node in the network.

	Time() func() int // Function that can be used to retrieve the simulation global virtual time.
}

type RoutineCapabilities interface {
	Routine(interval int, routine func()) Routine // allows setting up a periodic routine
	Callback(timeout int, routine func()) Callback // allows setting up a callback
}
```
