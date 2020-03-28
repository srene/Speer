package main

import (
	"fmt"
	"github.com/srene/Speer/discv5"
	"time"
)

func main() {
	/*rand.Seed(time.Now().UTC().UnixNano())

	// Parsing the flags
	flag.Parse()

	// Profiling
	defer makeCPUProfile()()
	defer makeMemprofile()
	setSignals()

	fmt.Println("Config "+*configPath)
	jsonConfig := config.JSONConfig(*configPath)
	simulation := config.NewSimulation(jsonConfig)
	simulation.Run()

	time.Sleep(time.Second * time.Duration(*secs))
	simulation.Stop()*/
	sim := discv5.NewSimulation()
	bootnode := sim.LaunchNode(false)

	launcher := time.NewTicker(10 * time.Second)
	go func() {
		fmt.Println("testing2")
		for range launcher.C {
			fmt.Println("testing3")
			net := sim.LaunchNode(true)
			go discv5.RandomResolves(sim, net)
			if err := net.SetFallbackNodes([]*discv5.Node{bootnode.Self()}); err != nil {
				panic(err)
			}
			fmt.Printf("launched @ %v: %x\n", time.Now(), net.Self().ID[:16])
		}
	}()

	time.Sleep(300 * time.Second)
	launcher.Stop()
	sim.Shutdown()
	sim.PrintStats()
}