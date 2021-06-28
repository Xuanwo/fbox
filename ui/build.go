package main

import "github.com/evanw/esbuild/pkg/api"
import "os"
import "fmt"

func main() {
  result := api.Build(api.BuildOptions{
    EntryPoints: []string{"src/index.js"},
    Outfile:     "index.min.js",
    Write:       true,
		Bundle:			 true,
		Watch: &api.WatchMode{
			OnRebuild: func(result api.BuildResult) {
				if len(result.Errors) > 0 {
					fmt.Printf("watch build failed: %d errors\n", len(result.Errors))
				} else {
					fmt.Printf("watch build succeeded: %d warnings\n", len(result.Warnings))
				}
			},
		},
  })

  if len(result.Errors) > 0 {
    os.Exit(1)
  }
}
