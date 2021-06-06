package blob

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/prologic/fbox/store"
)

type Handler struct {
	s store.Store
}

func NewHandler(s store.Store) http.Handler {
	return Handler{s}
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var logger *log.Entry

	status, body := func() (int, []byte) {
		hkey := r.URL.Path[1:]
		key, err := hex.DecodeString(hkey)
		if err != nil {
			return http.StatusBadRequest, []byte(
				fmt.Sprintf("%q: not a valid path, expecting hex key only",
					r.URL.Path),
			)
		}

		logger = log.WithFields(log.Fields{
			"op":  r.Method,
			"key": hkey,
		})

		switch r.Method {
		case http.MethodGet:
			value, err := h.s.Get(key)
			if errors.Is(err, store.ErrNotFound) {
				logger.WithField("err", err).Debug("Not found")
				return http.StatusNotFound, nil
			}
			if err != nil {
				logger.WithField("err", err).Error()
				return http.StatusInternalServerError, []byte(fmt.Sprintf("%q: %v", hkey, err))
			}
			logger.Debug("Success")
			return http.StatusOK, value
		case http.MethodPut:
			value, err := ioutil.ReadAll(r.Body)
			if err != nil {
				logger.WithField("err", err).Error()
				return http.StatusInternalServerError, []byte(fmt.Sprintf("%q: %v", hkey, err))
			}
			if err := h.s.Put(key, value); err != nil {
				logger.WithField("err", err).Error()
				return http.StatusInternalServerError, []byte(fmt.Sprintf("%q: %v", hkey, err))
			}
			logger.Debug("Success")
			return http.StatusOK, nil
		default:
			logger.Warn("Bad request")
			return http.StatusBadRequest, []byte(fmt.Sprintf("%q: invalid method, expecting GET or PUT", r.Method))
		}
	}()

	w.WriteHeader(status)
	if body != nil {
		if _, err := w.Write(body); err != nil {
			logger.WithField("err", err).Error("Failed writing response")
		}
	}
}
