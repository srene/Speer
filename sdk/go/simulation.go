package sdk

import (
	"github.com/srene/Speer/interfaces"

	"github.com/srene/Speer/capacity"
	"github.com/srene/Speer/events"
	"github.com/srene/Speer/overlay"
	"github.com/srene/Speer/underlay"

	"github.com/srene/Speer/logs"

	"fmt"
)

type ISimulation interface {
	interfaces.ISimulation
}

type Simulation struct {
	underlaySimulation *underlay.NetworkSimulation

	logger *logs.EventMonitor

	directMap   overlay.DirectMap
	latencyMap  overlay.LatencyMap
	capacityMap capacity.CapacityMap
	nodes       int
	cnode       int

	userNodes map[string]SpeerNode

	template interfaces.Node

	progressProperties []events.Receiver
}

type SimulationBuilder struct {
	*Simulation
}

func NewSimulationBuilder(template interfaces.Node) *SimulationBuilder {
	b := &SimulationBuilder{
		Simulation: new(Simulation),
	}

	if template == nil {
		panic("No valid template provided")
	}

	b.template = template
	b.progressProperties = []events.Receiver{}
	b.userNodes = map[string]SpeerNode{}
	b.nodes = -1

	// By default we consider there is no latency module present
	b.underlaySimulation = underlay.NewNetworkSimulation(
		events.NewLazySimulation(),
		nil,
	)
	b.directMap = overlay.NewChanMap()

	return b
}

func (b *SimulationBuilder) WithParallelSimulation() *SimulationBuilder {
	if b.underlaySimulation == nil {
		panic("Underlay simulation component has to be appended first.")
	}

	b.underlaySimulation.SetParallel(true)
	return b
}

func (b *SimulationBuilder) WithInternetworkUnderlay(
	transitDomains int,
	transitDomainSize int,
	stubDomains int,
	stubDomainSize int,
) *SimulationBuilder {
	network := underlay.NewInternetwork(
		transitDomains,
		transitDomainSize,
		stubDomains,
		stubDomainSize,
	)
	simulation := underlay.NewNetworkSimulation(
		events.NewLazySimulation(),
		network,
	)

	fmt.Printf("Internetwork built with %d nodes.\n", len(network.Routers))
	b.underlaySimulation = simulation
	b.latencyMap = overlay.NewNetworkMap(b.underlaySimulation.Network())
	b.directMap = nil

	return b
}

func (b *SimulationBuilder) WithRandomUniformUnderlay(
	nodes int,
	edges int,
	minLatency int,
	maxLatency int,
) *SimulationBuilder {
	b.underlaySimulation = underlay.NewNetworkSimulation(
		events.NewLazySimulation(),
		underlay.NewRandomUniformNetwork(
			nodes,
			edges,
			minLatency,
			maxLatency,
		),
	)
	b.latencyMap = overlay.NewNetworkMap(b.underlaySimulation.Network())
	b.directMap = nil

	return b
}

func (b *SimulationBuilder) WithFixedNodes(
	nodes int,
) *SimulationBuilder {
	b.nodes = nodes
	b.cnode = 0
	return b
}

func (b *SimulationBuilder) WithCapacityScheduler(
	interval int,
) *SimulationBuilder {
	b.capacityMap = capacity.NewScheduledCapacityMap(interval)
	return b
}

func (b *SimulationBuilder) addNewNode() (
	id string,
	controlConnector interfaces.ControlTransport,
	bootstrap overlay.Bootstrap,
	observer events.ActiveObserver,
) {
	if b.latencyMap != nil {
		// assign ID to node
		id = b.latencyMap.NewId()

		// create latency connector
		controlConnector = overlay.NewUnderlayChan(
			id,
			b.underlaySimulation,
			b.latencyMap,
		)
		observer = controlConnector.(overlay.LatencyConnector).Observer()

		bootstrap = b.latencyMap
	} else {
		// assign ID to node & create direct channel
		controlConnector, id = overlay.NewDirectChan(b.directMap)
		bootstrap = b.directMap
	}
	return id, controlConnector, bootstrap, observer
}

func (b *SimulationBuilder) WithCapacityNodes(
	nodes int,
	upload int,
	download int,
) *SimulationBuilder {
	if b.nodes == -1 {
		panic("Node number not specified.")
	}
	if b.capacityMap == nil {
		panic("No capacity scheduler provided.")
	}
	fmt.Println("Number of nodes %v %v %v \n",nodes,b.nodes,b.cnode)
	limit := b.cnode + nodes
	if b.nodes < limit {
		limit = b.nodes
	}
	for i := b.cnode; i < limit; i++ {
		id, controlConnector, bootstrap, latencyObserver := b.addNewNode()

		// register capacity
		capacityConnector := capacity.NewCapacityConnector(
			upload,
			download,
			b.capacityMap,
		)
		capacityObserver := events.NewActiveEventObserver(b.capacityMap.Receiver())
		b.capacityMap.AddConnector(id, capacityConnector)

		// register autowired node
		newNode := NewAutowiredNode(b.template, NewSimulatedNode(
			controlConnector,
			capacityConnector,
			b.underlaySimulation.Simulation,
			bootstrap,
			id,
			b.Time,
		))

		// set the observers for the autowired nodes
		if latencyObserver != nil {
			latencyObserver.SetProxy(events.NewProxy(newNode.OnNotify))
		}
		capacityObserver.SetProxy(events.NewProxy(newNode.OnNotify))
		b.underlaySimulation.RegisterObserver(capacityObserver)

		b.userNodes[id] = newNode
	}
	b.cnode = limit
	return b
}

func (b *SimulationBuilder) WithLogs(logsFile string) *SimulationBuilder {
	globalObserver := events.NewGlobalEventObserver()
	b.underlaySimulation.RegisterObserver(globalObserver)

	b.logger = logs.NewEventMonitor(globalObserver, b.latencyMap, logsFile)
	go b.logger.GatherEvents()

	return b
}

type looper struct {
	time func() int
}

func (l *looper) Receive(e *events.Event) *events.Event {
	return events.NewEvent(l.time()+1, nil, l)
}

func (b *SimulationBuilder) Build() ISimulation {
	if b.nodes == -1 {
		panic("Node number not specified.")
	}
	if b.underlaySimulation == nil {
		panic("No underlay simulation provided.")
	}

	looper := &looper{time: b.underlaySimulation.Time}
	if b.capacityMap == nil {
		for i := 0; i < b.nodes; i++ {
			id, controlConnector, bootstrap, latencyObserver := b.addNewNode()

			newNode := NewAutowiredNode(b.template, NewSimulatedNode(
				controlConnector,
				nil,
				b.underlaySimulation.Simulation,
				bootstrap,
				id,
				b.Time,
			))
			b.userNodes[id] = newNode

			// set the observers for the autowired nodes
			if latencyObserver != nil {
				latencyObserver.SetProxy(events.NewProxy(newNode.OnNotify))
			}
			if b.latencyMap == nil {
				// I have neither of the modules
				capacityObserver := events.NewActiveEventObserver(looper)
				capacityObserver.SetProxy(events.NewProxy(newNode.OnNotify))
				b.underlaySimulation.RegisterObserver(capacityObserver)
			}
		}
	}

	if b.capacityMap == nil && b.latencyMap == nil {
		b.underlaySimulation.Push(events.NewEvent(0, nil, looper))
	}

	return b.Simulation
}

func (s *Simulation) Run() {
	fmt.Println("Starting simulation at ",s.Time())
	for _, progress := range s.progressProperties {
		event := events.NewEvent(0, nil, progress)
		s.underlaySimulation.Push(event)
	}

	if s.capacityMap != nil {
		s.capacityMap.Start(s.underlaySimulation)
	}

	for _, node := range s.userNodes {
		if s.logger != nil {
			s.logger.Log(logs.JoinEntry{
				Time: s.Time(),
				Node: node.Id(),
			})
		}
		node.OnJoin()
	}
	go s.underlaySimulation.Run()
}

func (s *Simulation) Stop() {
	fmt.Println("Finishing simulation at ",s.Time())
	s.underlaySimulation.Stop()
}

func (s *Simulation) Time() int {
	return s.underlaySimulation.Time()
}
