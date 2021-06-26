package main

var (
	nodes []string
	files map[string]Metadata
)

func init() {
	files = make(map[string]Metadata)
}
