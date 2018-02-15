package events

import (
  "testing"
)

func TestMultipleTimeReads(t *testing.T) {
  /*
   * There is no current time ordering guarantee!
   * This is just to make sure there is a locking mechanism over the timer.
   */
  // s := NewLazySimulation()
  // r := new(mockReceiver)
  //
  // go s.Run()
  //
  // done := make(chan bool)
  //
  // for i := 1; i < LazyQueueChanSize; i++ {
  //   go func() {
  //     s.Push(NewEvent(i, nil, r))
  //     done <- true
  //     s.Time()
  //   }()
  // }
  //
  // for i := 1; i < LazyQueueChanSize; i++ {
  //   <-done
  // }
  //
  // s.Stop()
}

func TestObserversGetNotified(t *testing.T) {
  /*
   * There is no current time ordering guarantee!
   * Test observers get notified.
   */

  r1  := new(mockReceiver)
  r2  := new(mockReceiver)

  o := NewEventObserver(r1)
  s := NewLazySimulation()

  go s.Run()
  s.RegisterObserver(o)

  done := make(chan bool)

  go func() {
    for i := 1; i <= LazyQueueChanSize / 2; i++ {
      s.Push(NewEvent(i, nil, r1))
      done <- true
    }
  }()
  go func() {
    for i := 1; i <= LazyQueueChanSize / 2; i++ {
      s.Push(NewEvent(i, nil, r2))
      done <- true
    }
  }()

  for i := 1; i <= LazyQueueChanSize; i++ {
    <-done
  }

  for i := 1; i < LazyQueueChanSize/2; i++ {
    e := <- o.EventChan()
    if e.Timestamp() > LazyQueueChanSize/2 {
  		t.Fatalf("Inconsistent simulation times.")
  	}
    // For assserting time ordering
    // assertEqual(t, e.Timestamp(), i)
  }

  s.Stop()
}
