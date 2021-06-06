package main

type File struct {
	Name string
	Size int64
	Hash string

	Data   int
	Parity int
	Shards []string
}
