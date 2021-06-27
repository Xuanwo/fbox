package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/creasty/defaults"
	"github.com/prologic/bitcask"
	log "github.com/sirupsen/logrus"
)

var (
	nodes []string
	db    *bitcask.Bitcask
)

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

func (m *Metadata) Bytes() ([]byte, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func loadMetadata(data []byte) (metadata *Metadata, err error) {
	metadata = &Metadata{}
	if err := defaults.Set(metadata); err != nil {
		return nil, err
	}

	if err = json.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}

	return
}

func getAllMetadata() ([]*Metadata, error) {
	var ms []*Metadata

	err := db.Fold(func(key []byte) error {
		data, err := db.Get(key)
		if err != nil {
			return err
		}

		m, err := loadMetadata(data)
		if err != nil {
			return err
		}
		ms = append(ms, m)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return ms, nil
}

func getMetadata(key string) (*Metadata, bool, error) {
	val, err := db.Get([]byte(key))
	if err != nil {
		if errors.Is(err, bitcask.ErrKeyNotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}

	m, err := loadMetadata(val)
	if err != nil {
		return nil, false, err
	}
	return m, true, nil
}

func setMetadata(key string, m *Metadata) error {
	data, err := m.Bytes()
	if err != nil {
		log.WithError(err).Error("error serializing metdata")
		return fmt.Errorf("error serialzing metadata: %w", err)
	}

	if err := db.Put([]byte(key), data); err != nil {
		log.WithError(err).Error("error storing metdata")
		return fmt.Errorf("error storing metadata: %w", err)
	}

	return nil
}
