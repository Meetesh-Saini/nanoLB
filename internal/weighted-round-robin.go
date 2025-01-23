package lb

import (
	resetableonce "nanoLB/internal/resetableOnce"
	"sync/atomic"
)

var (
	weightedRoundRobin     *WeightedRoundRobin
	onceWeightedRoundRobin *resetableonce.ResettableOnce = resetableonce.NewResettableOnce()
)

type WeightedRoundRobin struct {
	current     uint64
	indexLookup []int64
}

func (r *WeightedRoundRobin) GetNext(sp *ServerPool) *Server {
	if len(sp.pool) == 0 {
		return nil
	}
	next := r.nextIndex(uint64(len(r.indexLookup)))
	l := len(r.indexLookup) + next
	for i := next; i < l; i++ {
		idx := i % len(r.indexLookup)
		sidx := r.indexLookup[idx]
		if sp.pool[sidx].IsHealthy() {
			if i != next {
				atomic.StoreUint64(&r.current, uint64(idx))
			}
			return sp.pool[sidx]
		}
	}
	return nil
}

func (r *WeightedRoundRobin) nextIndex(lookupLength uint64) int {
	return int(atomic.AddUint64(&r.current, uint64(1)) % lookupLength)
}

func (r *WeightedRoundRobin) MakeWeights(sp *ServerPool) {
	sp.mux.Lock()
	defer sp.mux.Unlock()

	if len(sp.pool) == 0 {
		return
	}

	total := int64(0)
	gcd := int64(0)
	for _, server := range sp.pool {
		total += server.Weight
		if gcd == 0 {
			gcd = server.Weight
		} else {
			gcd = _gcd(gcd, server.Weight)
		}
	}
	poolLength := total / gcd
	serverPoolLength := len(sp.pool)

	r.indexLookup = make([]int64, poolLength)

	// Smooth weighted round robin
	// Modified implementation of https://github.com/nginx/nginx/commit/52327e0627f49dbda1e8db695e63a4b0af4448b1
	current_weights := make([]int64, serverPoolLength)
	for i := int64(0); i < poolLength; i++ {
		// Add weight to current weight
		for j := range current_weights {
			current_weights[j] += sp.pool[j].Weight
		}

		// Select server with highest weight
		maxIdx := 0
		for i := range current_weights {
			if current_weights[i] > current_weights[maxIdx] {
				maxIdx = i
			}
		}

		// Reduce selected server's weight
		current_weights[maxIdx] -= total

		// Store the selected server in lookup
		r.indexLookup[i] = int64(maxIdx)
	}
}

func GetWeightedRoundRobin(pool *ServerPool) *WeightedRoundRobin {
	onceWeightedRoundRobin.Do(func() {
		weightedRoundRobin = &WeightedRoundRobin{current: ^uint64(0)}
		weightedRoundRobin.MakeWeights(pool)
	})
	return weightedRoundRobin
}
