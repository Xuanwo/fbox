package main

var (
	nodes []string
	files map[string]File
)

func init() {
	files = make(map[string]File)
}
