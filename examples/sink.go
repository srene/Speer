package examples

import (
	. "github.com/srene/Speer/interfaces"

	"fmt"
)

type SinkExample struct {
	Transport

	id     string
	parent string
	ctr    int
}

func (s *SinkExample) New(util NodeUtil) Node {
	fmt.println("New node"+util.Id())
	return &SinkExample{
		Transport: util.Transport(),

		id:     util.Id(),
		parent: util.Join(),
		ctr:    0,
	}
}

func (s *SinkExample) root() bool {
	return s.parent == ""
}

func (s *SinkExample) OnJoin() {
	// send my id to the parent
	if !s.root() {
		s.ControlSend(s.parent, s.id)
	}
}

func (s *SinkExample) OnNotify() {
	select {
	case m, _ := <-s.ControlRecv():
		if !s.root() {
			// forward each message
			s.ControlSend(s.parent, m)
		} else {
			// the root will print the messages
			s.ctr += 1
			fmt.Println("message #", s.ctr, "received", m)
		}
	default:
	}
}

func (s *SinkExample) OnLeave() {
}
