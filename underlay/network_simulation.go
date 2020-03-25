package underlay

import (
	. "github.com/srene/Speer/events"
)

// A NetworkSimulation is a Simulation with a Network attached.
type NetworkSimulation struct {
	*Simulation
	network *Network
}

func NewNetworkSimulation(s *Simulation, n *Network) *NetworkSimulation {
	return &NetworkSimulation{
		Simulation: s,
		network:    n,
	}
}

func (s *NetworkSimulation) SendPacket(p Packet) {
	s.Push(NewEvent(s.Time(), p, p.Src()))
}

func (s *NetworkSimulation) Network() *Network {
	return s.network
}
