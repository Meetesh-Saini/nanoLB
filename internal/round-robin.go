package lb

import (
	resetableonce "nanoLB/internal/resetableOnce"
	"sync/atomic"
)

var (
	roundRobin     *RoundRobin
	onceRoundRobin *resetableonce.ResettableOnce = resetableonce.NewResettableOnce()
)

type RoundRobin struct {
	current uint64
}

func (r *RoundRobin) GetNext(sp *ServerPool) *Server {
	if len(sp.pool) == 0 {
		return nil
	}
	next := r.nextIndex(uint64(len(sp.pool)))
	l := len(sp.pool) + next
	for i := next; i < l; i++ {
		idx := i % len(sp.pool)
		if sp.pool[idx].IsHealthy() {
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
		roundRobin = &RoundRobin{current: ^uint64(0)}
	})
	return roundRobin
}
