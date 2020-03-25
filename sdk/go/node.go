package sdk

import (
	"github.com/srene/Speer/interfaces"
)

type SpeerNode interface {
	interfaces.Node
	interfaces.NodeUtil
}

type AutowiredNode struct {
	interfaces.Node
	interfaces.NodeUtil
}

func NewAutowiredNode(
	template interfaces.Node,
	util interfaces.NodeUtil,
) SpeerNode {
	return &AutowiredNode{
		Node:     template.New(util),
		NodeUtil: util,
	}
}
