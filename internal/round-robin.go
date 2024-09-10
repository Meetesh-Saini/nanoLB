package lb

import (
	"sync"
	"sync/atomic"
)

var (
	roundRobin     *RoundRobin
	onceRoundRobin sync.Once
)

type RoundRobin struct {
	current uint64
}

func (r *RoundRobin) GetNext(sp *ServerPool) *Server {
	if len(sp.pool) == 0 {
		return nil
	}
	next := r.nextIndex(uint64(len(sp.pool)))
	l := len(sp.pool) + next // start from next and move a full cycle
	for i := next; i < l; i++ {
		idx := i % len(sp.pool)       // take an index by modding
		if sp.pool[idx].IsHealthy() { // if we have an alive backend, use it and store if its not the original one
			if i != next {
				atomic.StoreUint64(&r.current, uint64(idx))
			}
			return sp.pool[idx]
		}
	}
	return nil
}

func (r *RoundRobin) nextIndex(poolLen uint64) int {
	return int(atomic.AddUint64(&r.current, uint64(1)) % poolLen)
}

func GetRoundRobin() *RoundRobin {
	onceRoundRobin.Do(func() {
		roundRobin = &RoundRobin{current: 0}
	})
	return roundRobin
}
