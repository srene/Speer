package main

import (
	"time"
	"github.com/srene/Speer/discv5"

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

	sim := newSimulation()
	bootnode := sim.launchNode(false)

	launcher := time.NewTicker(10 * time.Second)
	go func() {
		fmt.Println("testing2")
		for range launcher.C {
			fmt.Println("testing3")
			net := sim.launchNode(true)
			go randomResolves(sim, net)
			if err := net.SetFallbackNodes([]*Node{bootnode.Self()}); err != nil {
				panic(err)
			}
			fmt.Printf("launched @ %v: %x\n", time.Now(), net.Self().ID[:16])
		}
	}()

	time.Sleep(300 * time.Second)
	launcher.Stop()
	sim.shutdown()
	sim.printStats()
}