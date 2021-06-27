package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/tomasen/realip"
	"github.com/unrolled/render"
)

var r *render.Render

func init() {
	r = render.New()
}

func joinHandler(w http.ResponseWriter, req *http.Request) {
	data := map[string]string{}
	if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
		msg := fmt.Sprintf("error decoding join request: %s", err)
		log.WithError(err).Error(msg)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	remoteAddr, ok := data["addr"]
	if !ok {
		remoteAddr = realip.FromRequest(req)
	}

	addr, err := net.ResolveTCPAddr("tcp4", remoteAddr)
	if err != nil {
		msg := fmt.Sprintf("error resolving remoteAddr %s: %s", remoteAddr, err)
		log.WithError(err).Error(msg)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if addr.IP == nil {
		reqAddr, err := net.ResolveTCPAddr("tcp4", req.RemoteAddr)
		if err != nil {
			msg := fmt.Sprintf("error resolving reqAddr %s: %s", req.RemoteAddr, err)
			log.WithError(err).Error(msg)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		addr.IP = reqAddr.IP
	}

	if addr.Port == 0 {
		// TODO: Make this a constant
		addr.Port = 8000
	}

	remoteAddr = addr.String()

	log.Infof("node joined from %s", remoteAddr)
	nodes = append(nodes, remoteAddr)
}

func nodesHandler(w http.ResponseWriter, req *http.Request) {
	r.JSON(w, http.StatusOK, nodes)
}

func filesHandler(w http.ResponseWriter, req *http.Request) {
	files, err := getAllMetadata()
	if err != nil {
		msg := fmt.Sprintf("error reading all metadata: %s", err)
		log.WithError(err).Error(msg)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	r.JSON(w, http.StatusOK, files)
}

func metadataHandler(w http.ResponseWriter, req *http.Request) {
	name := req.URL.Path

	metadata, ok, err := getMetadata(name)
	if err != nil {
		msg := fmt.Sprintf("error getting metdata for %s: %s", name, err)
		log.Error(msg)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !ok {
		msg := fmt.Sprintf("error file not found: %s", name)
		log.Error(msg)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	r.JSON(w, http.StatusOK, metadata)
}

func uploadHandler(w http.ResponseWriter, req *http.Request) {
	name := req.URL.Path

	tf, err := receiveFile(req.Body)
	if err != nil {
		msg := fmt.Sprintf("error receiving file: %s", err)
		log.WithError(err).Error(msg)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	hash, err := hashReader(tf)
	if err != nil {
		msg := fmt.Sprintf("error hashing file: %s", err)
		log.WithError(err).Error(msg)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if _, err := tf.Seek(0, io.SeekStart); err != nil {
		msg := fmt.Sprintf("error seeking file: %s", err)
		log.WithError(err).Error(msg)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	stat, err := os.Stat(tf.Name())
	if err != nil {
		msg := fmt.Sprintf("error getting file size: %s", err)
		log.WithError(err).Error(msg)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	shards, err := createShards(tf, stat.Size())
	if err != nil {
		msg := fmt.Sprintf("error creating shards: %s", err)
		log.WithError(err).Error(msg)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	uris, err := storeShards(shards)
	if err != nil {
		msg := fmt.Sprintf("error storing shards: %s", err)
		log.WithError(err).Error(msg)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := setMetadata(name, &Metadata{
		Name:   name,
		Size:   stat.Size(),
		Hash:   hash,
		Parity: parityShards,
		Shards: uris,
	}); err != nil {
		msg := fmt.Sprintf("error storing metadata: %s", err)
		log.WithError(err).Error(msg)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func downloadHandler(w http.ResponseWriter, req *http.Request) {
	name := req.URL.Path

	metadata, ok, err := getMetadata(name)
	if err != nil {
		msg := fmt.Sprintf("error getting metdata for %s: %s", name, err)
		log.Error(msg)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !ok {
		msg := fmt.Sprintf("error file not found: %s", name)
		log.Error(msg)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	f, err := readShards(metadata)
	if err != nil {
		msg := fmt.Sprintf("error reading file: %s", err)
		log.WithError(err).Error(msg)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := io.Copy(w, f); err != nil {
		msg := fmt.Sprintf("error sending file: %s", err)
		log.WithError(err).Error(msg)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
