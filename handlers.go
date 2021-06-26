package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
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
	r.JSON(w, http.StatusOK, files)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}

	log.Debugf("req: %q", dump)

	name := r.URL.Path

	tf, err := receiveFile(r.Body)
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

	files[name] = Metadata{
		Name:   name,
		Size:   stat.Size(),
		Hash:   hash,
		Shards: uris,
	}
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {}
