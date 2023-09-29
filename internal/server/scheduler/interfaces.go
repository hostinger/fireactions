package scheduler

import (
	"github.com/hostinger/fireactions/internal/structs"
)

type Cache interface {
	GetNodes() ([]*structs.Node, error)
	AddNode(n *structs.Node) error
	DelNode(n *structs.Node) error
	PutNode(n *structs.Node) error
	DeepCopy() interface{}
}
