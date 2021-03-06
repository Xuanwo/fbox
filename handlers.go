package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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

func metadataHandler(w http.ResponseWriter, r *http.Request) {
	var logger *log.Entry

	status, body := func() (int, []byte) {
		name := r.URL.Path

		logger = log.WithFields(log.Fields{
			"op":   r.Method,
			"name": name,
		})

		switch r.Method {
		case http.MethodDelete:
			ok, err := deleteMetadata(name)
			if err != nil {
				logger.WithField("err", err).Error()
				return http.StatusInternalServerError, []byte(fmt.Sprintf("%q: %v", name, err))
			}
			if !ok {
				logger.WithField("err", err).Debug("Not found")
				return http.StatusNotFound, nil
			}
			logger.Debug("Success")
			return http.StatusOK, nil
		case http.MethodGet:
			metadata, ok, err := getMetadata(name)
			if err != nil {
				logger.WithField("err", err).Error()
				return http.StatusInternalServerError, []byte(fmt.Sprintf("%q: %v", name, err))
			}

			if !ok {
				logger.WithField("err", err).Debug("Not found")
				return http.StatusNotFound, nil
			}

			data, err := metadata.Bytes()
			if err != nil {
				logger.WithField("err", err).Error()
				return http.StatusInternalServerError, []byte(fmt.Sprintf("%q: %v", name, err))
			}
			logger.Debug("Success")
			return http.StatusOK, data
		case http.MethodPost:
			value, err := ioutil.ReadAll(r.Body)
			if err != nil {
				logger.WithField("err", err).Error()
				return http.StatusInternalServerError, []byte(fmt.Sprintf("%q: %v", name, err))
			}
			metadata, err := loadMetadata(value)
			if err != nil {
				logger.WithField("err", err).Error()
				return http.StatusInternalServerError, []byte(fmt.Sprintf("%q: %v", name, err))
			}

			if err := setMetadata(name, metadata); err != nil {
				logger.WithField("err", err).Error()
				return http.StatusInternalServerError, []byte(fmt.Sprintf("%q: %v", name, err))
			}
			logger.Debug("Success")
			return http.StatusOK, nil
		default:
			logger.Warn("Bad request")
			return http.StatusBadRequest, []byte(
				fmt.Sprintf("%q: invalid method, expecting DELETE, GET,  or PUT",
					r.Method,
				))
		}
	}()

	w.WriteHeader(status)
	if body != nil {
		if _, err := w.Write(body); err != nil {
			logger.WithField("err", err).Error("Failed writing response")
		}
	}

}

func uploadHandler(w http.ResponseWriter, req *http.Request) {
	name := req.URL.Path
	if len(name) > maxPathLength {
		msg := fmt.Sprintf("error name %d length exceeds allowed maximum of %d", len(name), maxPathLength)
		log.Warn(msg)
		http.Error(w, msg, http.StatusUnprocessableEntity)
		return
	}

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
