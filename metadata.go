package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

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

func getRemoteMetadata(addr string, name string) (*Metadata, bool, error) {
	uri := fmt.Sprintf("http://%s/metadata/%s", addr, name)
	res, err := request(http.MethodGet, uri, nil, nil)
	if err != nil {
		log.WithError(err).Error("error making metdata request")
		return nil, false, fmt.Errorf("error making metadata request: %w", err)
	}
	if res.StatusCode != 200 {
		log.WithField("Status", res.Status).Error("error making metadata request")
		return nil, false, fmt.Errorf("error making metadata request: %s", res.Status)
	}
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.WithError(err).Error("error reading metadata response")
		return nil, false, fmt.Errorf("error reading metadata response: %w", err)
	}

	m, err := loadMetadata(data)
	if err != nil {
		return nil, false, err
	}
	return m, true, nil
}

func setRemoteMetadata(addr string, name string, metadata *Metadata) error {
	data, err := metadata.Bytes()
	if err != nil {
		log.WithError(err).Error("error serializing metdata")
		return fmt.Errorf("error serializing metadata: %w", err)
	}

	uri := fmt.Sprintf("http://%s/metadata/%s", addr, name)
	res, err := request(http.MethodPost, uri, nil, bytes.NewReader(data))
	if err != nil {
		log.WithError(err).Error("error making metdata request")
		return fmt.Errorf("error making metadata request: %w", err)
	}
	if res.StatusCode != 200 {
		log.WithField("Status", res.Status).Error("error making metadata request")
		return fmt.Errorf("error making metadata request: %s", res.Status)
	}
	defer res.Body.Close()

	return nil
}

func deleteMetadata(key string) (bool, error) {
	if !db.Has([]byte(key)) {
		return false, nil
	}

	if err := db.Delete([]byte(key)); err != nil {
		return false, err
	}

	return true, nil
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
