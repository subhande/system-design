package main

import (
	"fmt"
)

type BalacingStrategy interface {
	Init([]*Backend)
	GetNextBackend(IncomingReq) *Backend
	RegisterBackend(*Backend)
	PrintTopology()
}

type RRBalancingStrategy struct {
	Index    int
	Backends []*Backend
}

type StaticBalancingStrategy struct {
	Index    int
	Backends []*Backend
}

type HashedBalancingStrategy struct {
	OccupiedSlots []int
	Backends      []*Backend
}

func (s *RRBalancingStrategy) Init(backends []*Backend) {
	s.Index = 0
	s.Backends = backends
}

func (s *RRBalancingStrategy) GetNextBackend(req IncomingReq) *Backend {
	s.Index = (s.Index + 1) % len(s.Backends)
	return s.Backends[s.Index]
}

func (s *RRBalancingStrategy) RegisterBackend(b *Backend) {
	s.Backends = append(s.Backends, b)
}

func (s *RRBalancingStrategy) PrintTopology() {
	for _, b := range s.Backends {
		fmt.Println(b)
	}
}

func (s *StaticBalancingStrategy) Init(backends []*Backend) {
	s.Index = 0
	s.Backends = backends
}

func (s *StaticBalancingStrategy) GetNextBackend(req IncomingReq) *Backend {
	return s.Backends[s.Index]
}

func (s *StaticBalancingStrategy) RegisterBackend(b *Backend) {
	s.Backends = append(s.Backends, b)
}

func (s *StaticBalancingStrategy) PrintTopology() {
	for _, b := range s.Backends {
		fmt.Println(b)
	}
}

func (s *HashedBalancingStrategy) Init(backends []*Backend) {
	s.OccupiedSlots = make([]int, len(backends))
	s.Backends = backends
}

// func (s *HashedBalancingStrategy) GetNextBackend(req IncomingReq) *Backend {
// 	slot := hash(req.reqId)
// 	index := sort.Search(len(s.OccupiedSlots), func(i int) bool { return s.OccupiedSlots[i] >= slot })
// 	return s.Backends[index%len(s.Backends)]
// }

// func (s *HashedBalancingStrategy) RegisterBackend(b *Backend) {
// 	s.Backends = append(s.Backends, b)
// }

// func (s *HashedBalancingStrategy) PrintTopology() {
// 	for _, b := range s.Backends {
// 		fmt.Println(b)
// 	}
// }

// func NewHashedBalancingStrategy(backends []*Backend) *HashedBalancingStrategy {
// 	s := &HashedBalancingStrategy{}
// 	s.Init(backends)
// 	return s
// }

func NewRoundRobinBalancingStrategy(backends []*Backend) *RRBalancingStrategy {
	s := &RRBalancingStrategy{}
	s.Init(backends)
	return s
}
