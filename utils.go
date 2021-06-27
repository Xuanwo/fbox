package main

import (
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/klauspost/reedsolomon"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/blake2b"
)

const (
	requestTimeout = time.Second * 30
	dataShards     = 3
	parityShards   = 1
)

func receiveFile(r io.Reader) (*os.File, error) {
	tf, err := ioutil.TempFile("", "fbox-receive-*")
	if err != nil {
		log.WithError(err).Error("error creating temporary file")
		return nil, err
	}

	if _, err := io.Copy(tf, r); err != nil {
		log.WithError(err).Error("error writing temporary file")
		return tf, err
	}

	if _, err := tf.Seek(0, io.SeekStart); err != nil {
		log.WithError(err).Error("error seeking temporary file")
		return tf, err
	}

	return tf, nil
}

func createShards(r io.ReadSeeker, size int64) ([]io.ReadSeeker, error) {
	enc, err := reedsolomon.NewStream(dataShards, parityShards)
	if err != nil {
		log.WithError(err).Error("error creating reedsolomon stream")
		return nil, fmt.Errorf("error creating reedsolomon stream: %w", err)
	}

	shards := dataShards + parityShards
	out := make([]*os.File, shards)

	// Create the resulting files.
	for i := range out {
		tf, err := ioutil.TempFile("", "fbox-receive-*")
		if err != nil {
			log.WithError(err).Error("error creating temporary output file")
			return nil, err
		}
		out[i] = tf
	}

	// Split into files.
	data := make([]io.Writer, dataShards)
	for i := range data {
		data[i] = out[i]
	}

	// Do the split
	if err := enc.Split(r, data, size); err != nil {
		log.WithError(err).Error("error splitting input")
		return nil, fmt.Errorf("error splitting input: %w", err)
	}

	// Close and re-open the files.
	input := make([]io.Reader, dataShards)

	for i := range data {
		if err := out[i].Close(); err != nil {
			log.WithError(err).Error("error closing output")
			return nil, fmt.Errorf("error closing output: %w", err)
		}

		f, err := os.Open(out[i].Name())
		if err != nil {
			log.WithError(err).Error("error reopening output")
			return nil, fmt.Errorf("error reopening output: %w", err)
		}
		defer f.Close()
		input[i] = f
	}

	// Create parity output writers
	parity := make([]io.Writer, parityShards)
	for i := range parity {
		parity[i] = out[dataShards+i]
		defer out[dataShards+i].Close()
	}

	// Encode parity
	if err := enc.Encode(input, parity); err != nil {
		log.WithError(err).Error("error encoding party shards")
		return nil, fmt.Errorf("error encoding parity shards: %w", err)
	}

	// Close and reopen outputs and return as slice of readers
	rs := make([]io.ReadSeeker, shards)
	for i := range out {
		_ = out[i].Close()
		f, err := os.Open(out[i].Name())
		if err != nil {
			log.WithError(err).Error("error reopening output")
			return nil, fmt.Errorf("error reopening output: %w", err)
		}
		rs[i] = f
	}

	return rs, nil
}

func getShard(uri string) (io.Reader, error) {
	tf, err := ioutil.TempFile("", "fbox-get-*")
	if err != nil {
		log.WithError(err).Error("error creating temporary shard file")
		return nil, err
	}

	res, err := request(http.MethodGet, uri, nil, nil)
	if err != nil {
		log.WithError(err).Error("error making shard request")
		return nil, fmt.Errorf("error making shard request: %w", err)
	}

	if res.StatusCode != 200 {
		log.WithField("Status", res.Status).Error("non-200 shard response")
		return nil, fmt.Errorf("non-200 shard response")
	}
	defer res.Body.Close()

	if _, err := io.Copy(tf, res.Body); err != nil {
		log.WithError(err).Error("error reading shard")
		return nil, fmt.Errorf("error reading sahrd: %w", err)
	}

	if err := tf.Close(); err != nil {
		log.WithError(err).Error("error closing tempoary shard file")
		return nil, fmt.Errorf("error closing temporary shard file: %w", err)
	}

	return os.Open(tf.Name())
}

func getShards(uris []string) ([]io.Reader, error) {
	shards := make([]io.Reader, len(uris))

	for i, uri := range uris {
		shard, err := getShard(uri)
		if err != nil {
			log.WithError(err).WithField("uri", uri).Error("error getting shard")
			shards[i] = nil
			continue
		}
		shards[i] = shard
	}

	return shards, nil
}

func repairShards(enc reedsolomon.StreamEncoder, shards []io.Reader) ([]io.Reader, error) {
	// Create out destination writers
	out := make([]io.Writer, len(shards))
	for i := range out {
		if shards[i] == nil {
			tf, err := ioutil.TempFile("", "fbox-repair-*")
			if err != nil {
				log.WithError(err).Error("error creating temporary shard file")
				return nil, err
			}

			out[i] = tf
		}
	}

	if err := enc.Reconstruct(shards, out); err != nil {
		log.WithError(err).Error("error reconstructing shards")
		return nil, fmt.Errorf("error reconstructing shards: %w", err)
	}

	// Close output.
	for i := range out {
		if out[i] != nil {
			if err := out[i].(*os.File).Close(); err != nil {
				log.WithError(err).Error("error closing shard file")
				return nil, fmt.Errorf("error closing shard file: %w", err)
			}
			f, err := os.Open(out[i].(*os.File).Name())
			if err != nil {
				log.WithError(err).Error("error reopening repaired sshard file")
				return nil, fmt.Errorf("error reopening reparied shard file: %w", err)
			}
			shards[i] = f
		}
	}

	// Reset shards so we can re-read
	for _, shard := range shards {
		_, _ = shard.(*os.File).Seek(0, io.SeekStart)
	}

	ok, err := enc.Verify(shards)
	if err != nil {
		log.WithError(err).Error("error verifying repaired shards")
		return nil, fmt.Errorf("error verifying repaired shards: %w", err)
	}

	if !ok {
		log.Error("error verifying repaired shards")
		return nil, fmt.Errorf("error verifying repaired shards")
	}

	return shards, nil
}

func readShards(metadata *Metadata) (io.Reader, error) {
	// Create matrix
	enc, err := reedsolomon.NewStream(dataShards, parityShards)
	if err != nil {
		log.WithError(err).Error("error creating reedsolomon stream")
		return nil, fmt.Errorf("error creating reedsolomon stream: %w", err)
	}

	// Open the inputs
	shards, err := getShards(metadata.Shards)
	if err != nil {
		log.WithError(err).Error("error getting shards")
		return nil, fmt.Errorf("error getting shards: %s", err)
	}

	// Verify the shards
	ok, err := enc.Verify(shards)
	if err != nil {
		log.WithError(err).Warn("error verifying shards")
	}
	if !ok {
		log.Warn("shard verification failed, reconstructing shards...")

		// Reset shards so we can re-read
		for _, shard := range shards {
			if shard != nil {
				_, _ = shard.(*os.File).Seek(0, io.SeekStart)
			}
		}

		shards, err = repairShards(enc, shards)
		if err != nil {
			log.WithError(err).Error("error repairing shards")
			return nil, fmt.Errorf("error repairing shards: %s", err)
		}
	}

	// Reset shards so we can re-read
	for _, shard := range shards {
		_, _ = shard.(*os.File).Seek(0, io.SeekStart)
	}

	tf, err := ioutil.TempFile("", "fbox-*")
	if err != nil {
		log.WithError(err).Error("error creating temporary file")
		return nil, err
	}

	// Join the shards and write them
	if err := enc.Join(tf, shards, metadata.Size); err != nil {
		log.WithError(err).Error("error joining shards")
		return nil, fmt.Errorf("error joining shards: %w", err)
	}

	// Reset output file so we can re-read
	_, _ = tf.Seek(0, io.SeekStart)

	return tf, nil
}

// TODO: Add other node selection algorithms
//       For example: affinity+random selection
func selectNode(nodes []string) string {
	n := rand.Int() % len(nodes)
	return nodes[n]
}

func storeShards(rs []io.ReadSeeker) ([]string, error) {
	hashes, err := hashShards(rs)
	if err != nil {
		log.WithError(err).Error("error calculating shard hashses")
		return nil, fmt.Errorf("error calculating shard hashes: %w", err)
	}

	uris := make([]string, len(rs))
	for i, _ := range uris {
		uris[i] = fmt.Sprintf("http://%s/blob/%s", selectNode(nodes), hashes[i])
	}

	for i, uri := range uris {
		res, err := request(http.MethodPut, uri, nil, rs[i])
		if err != nil {
			log.WithError(err).Error("error making blob request")
			return nil, fmt.Errorf("error making blob request: %w", err)
		}
		if res.StatusCode != 200 {
			log.WithField("Status", res.Status).Error("error making blob request")
			return nil, fmt.Errorf("error making blob request: %s", res.Status)
		}
		defer res.Body.Close()
	}

	return uris, nil
}

func hashShards(rs []io.ReadSeeker) ([]string, error) {
	hashes := make([]string, len(rs))
	for i, r := range rs {
		hash, err := hashReader(r)
		if err != nil {
			log.WithError(err).Error("error hashing shard")
			return nil, fmt.Errorf("error hashing shard: %w", err)
		}
		if _, err := r.Seek(0, io.SeekStart); err != nil {
			log.WithError(err).Error("error seeking shard")
			return nil, fmt.Errorf("error seeking shard: %w", err)
		}

		hashes[i] = hash
	}

	return hashes, nil
}

func hashReader(r io.Reader) (string, error) {
	hasher, err := blake2b.New256(nil)
	if err != nil {
		log.WithError(err).Error("error creating hasher interface")
		return "", fmt.Errorf("error creating hasher interface: %s", err)
	}

	if _, err := io.Copy(hasher, r); err != nil {
		log.WithError(err).Error("error hashing reader")
		return "", fmt.Errorf("error hashing reader: %w", err)
	}

	sum := hasher.Sum(nil)

	return hex.EncodeToString(sum), nil
}

func request(method, url string, headers http.Header, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		log.WithError(err).Errorf("%s: http.NewRequest fail: %s", url, err)
		return nil, err
	}

	if headers == nil {
		headers = make(http.Header)
	}

	// Set a default User-Agent (if none set)
	if headers.Get("User-Agent") == "" {
		headers.Set("User-Agent", fmt.Sprintf("fbox/%s", FullVersion()))
	}

	req.Header = headers

	client := http.Client{
		Timeout: requestTimeout,
	}

	res, err := client.Do(req)
	if err != nil {
		log.WithError(err).Errorf("%s: client.Do fail: %s", url, err)
		return nil, err
	}

	return res, nil
}

func resourceExists(url string) bool {
	res, err := request(http.MethodHead, url, nil, nil)
	if err != nil {
		log.WithError(err).Errorf("error checking if %s exists", url)
		return false
	}
	defer res.Body.Close()

	return res.StatusCode/100 == 2
}

func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
