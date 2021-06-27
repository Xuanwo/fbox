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

func cmdPut(mAddr string, args []string) (int, error) {
	if len(args) != 1 {
		log.Errorf("incorrect no. of arguments, expected 1 got %d", len(args))
		return 1, fmt.Errorf("incorrect no. of arguments, expected 1 got %d", len(args))
	}

	name := args[0]

	if err := refreshNodes(mAddr); err != nil {
		log.WithError(err).Errorf("error refreshing nodes")
		return 2, fmt.Errorf("error refreshing nodes: %w", err)
	}

	f, err := os.Open(name)
	if err != nil {
		log.WithError(err).Errorf("error opening file: %s", name)
		return 2, fmt.Errorf("error opening file %s: %w", name, err)
	}
	defer f.Close()

	hash, err := hashReader(f)
	if err != nil {
		log.WithError(err).Errorf("error hashing file: %s", name)
		return 2, fmt.Errorf("error hashing file %s: %w", name, err)
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		log.WithError(err).Error("error seeking file")
		return 2, fmt.Errorf("error seeking file: %w", err)
	}

	stat, err := os.Stat(f.Name())
	if err != nil {
		log.WithError(err).Errorf("error getting file size")
		return 2, fmt.Errorf("error getting file size: %w", err)
	}

	shards, err := createShards(f, stat.Size())
	if err != nil {
		log.WithError(err).Errorf("error creating shards")
		return 2, fmt.Errorf("error creating shards: %w", err)
	}

	uris, err := storeShards(shards)
	if err != nil {
		log.WithError(err).Errorf("error storing shards")
		return 2, fmt.Errorf("error storing shards: %w", err)
	}

	if err := setRemoteMetadata(mAddr, name, &Metadata{
		Name:   name,
		Size:   stat.Size(),
		Hash:   hash,
		Parity: parityShards,
		Shards: uris,
	}); err != nil {
		log.WithError(err).Errorf("error storing metadata")
		return 2, fmt.Errorf("error storing metadata: %w", err)
	}

	return 0, nil
}
