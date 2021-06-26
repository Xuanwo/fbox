package main

type Metadata struct {
	Name   string
	Size   int64
	Hash   string
	Parity int
	Shards []string
}

func (m Metadata) DataShards() int {
	return len(m.Shards) - m.Parity
}

func (m Metadata) ParityShards() int {
	return m.Parity
}
