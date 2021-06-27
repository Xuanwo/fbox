package main

import (
	"fmt"
	"io"
	"os"

	log "github.com/sirupsen/logrus"
)

func cmdCat(mAddr string, args []string) (int, error) {
	if len(args) != 1 {
		log.Errorf("incorrect no. of arguments, expected 1 got %d", len(args))
		return 1, fmt.Errorf("incorrect no. of arguments, expected 1 got %d", len(args))
	}

	name := args[0]

	metadata, ok, err := getRemoteMetadata(mAddr, name)
	if err != nil {
		log.WithError(err).Errorf("error getting metdata for %s", name)
		return 2, fmt.Errorf("error getting metdata for %s: %w", name, err)
	}

	if !ok {
		log.WithError(err).Errorf("file not found %s", name)
		return 2, fmt.Errorf("file not found %s", name)
	}

	f, err := readShards(metadata)
	if err != nil {
		log.WithError(err).Errorf("error reading file")
		return 2, fmt.Errorf("error reading file: %w", err)
	}

	if _, err := io.Copy(os.Stdout, f); err != nil {
		log.WithError(err).Errorf("error writing file %s to stdout", name)
		return 2, fmt.Errorf("error writing file %s to stdout: %s", name, err)
	}

	return 0, nil
}
