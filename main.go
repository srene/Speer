package main

import (
  . "github.com/danalex97/Speer/sdk/go"
  . "github.com/danalex97/Speer/examples"

  errLog "log"
  "runtime/pprof"
  "runtime"
  "os/signal"
  "flag"

  "math/rand"
  "time"
  "fmt"
  "os"
)

var cpuprofile = flag.String("cpuprofile", "", "Write cpu profile to `file`.")
var memprofile = flag.String("memprofile", "", "Write memory profile to `file`.")

func makeMemprofile() {
  // Profiling
  if *memprofile != "" {
    f, err := os.Create(*memprofile)
    if err != nil {
        errLog.Fatal("could not create memory profile: ", err)
    }
    runtime.GC() // get up-to-date statistics
    if err := pprof.WriteHeapProfile(f); err != nil {
        errLog.Fatal("could not write memory profile: ", err)
    }
    f.Close()
  }
}

func main() {
  rand.Seed(time.Now().UTC().UnixNano())

  // Parsing the flags
  flag.Parse()

  // Profiling
  if *cpuprofile != "" {
    f, err := os.Create(*cpuprofile)
    if err != nil {
        errLog.Fatal("could not create CPU profile: ", err)
    }
    if err := pprof.StartCPUProfile(f); err != nil {
        errLog.Fatal("could not start CPU profile: ", err)
    }
    defer pprof.StopCPUProfile()
  }
  defer makeMemprofile()
  // Get profile even on signal
  c := make(chan os.Signal, 1)
  signal.Notify(c, os.Interrupt)
  go func(){
    for sig := range c {
      fmt.Println("Singal received:", sig)
      if *cpuprofile != "" {
        pprof.StopCPUProfile()
      }
      makeMemprofile()

      os.Exit(0)
    }
  }()


  nodeTemplate := new(SimpleTorrent)
  // nodeTemplate := new(SimpleTree)
  s := NewDHTSimulationBuilder(nodeTemplate).
    WithPoissonProcessModel(2, 2).
    // WithRandomUniformUnderlay(1000, 5000, 2, 10).
    WithInternetworkUnderlay(10, 20, 20, 50).
    // WithParallelSimulation().
    // WithInternetworkUnderlay(10, 50, 100, 100).
    WithDefaultQueryGenerator().
    WithLimitedNodes(100).
    // WithMetrics().
    //====================================
    WithCapacities().
    WithLatency().
    WithTransferInterval(10).
    WithCapacityNodes(100, 10, 20).
    WithCapacityNodes(100, 30, 30).
    Build()

  s.Run()

  time.Sleep(time.Second * 100)
  fmt.Println("Done")
  s.Stop()

  os.Exit(0)
}
