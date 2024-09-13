package config

type ALGORITHM_TYPE int

const (
	RoundRobin ALGORITHM_TYPE = iota
	WeightedRoundRobin
)
