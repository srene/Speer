package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"github.com/srene/Speer/discv5"
	"os"
	"time"
)

func main() {
	/*rand.Seed(time.Now().UTC().UnixNano())

	// Parsing the flags
	flag.Parse()

	// Profiling
	defer makeCPUProfile()()
	defer ma	keMemprofile()
	setSignals()

	fmt.Println("Config "+*configPath)
	jsonConfig := config.JSONConfig(*configPath)
	simulation := config.NewSimulation(jsonConfig)
	simulation.Run()

	time.Sleep(time.Second * time.Duration(*secs))
	simulation.Stop()*/
	sim := discv5.NewSimulation()
	bootnode := sim.LaunchNode(true)

	log.Root().SetHandler(log.LvlFilterHandler(log.LvlTrace, log.StreamHandler(os.Stderr, log.TerminalFormat(false))))

	fmt.Printf("Boot node %x \n",bootnode.Self().ID[:16])
	launcher := time.NewTicker(10 * time.Second)
	go func() {
		for range launcher.C {
			net := sim.LaunchNode(true)
			fmt.Printf("Launching new Node %x \n",net.Self().ID[:16])
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