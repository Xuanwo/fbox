package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

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

func uploadHandler(w http.ResponseWriter, r *http.Request)   {}
func downloadHandler(w http.ResponseWriter, r *http.Request) {}
