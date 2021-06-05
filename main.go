package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"github.com/unrolled/logger"

	"github.com/prologic/fbox/store"
)

var (
	debug   bool
	version bool

	bind string
	dir  string
)

const helpText = `
fbox is a simple distributed file system...

Valid options:
`

func init() {
	baseProg := filepath.Base(os.Args[0])
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", baseProg)
		fmt.Fprint(os.Stderr, helpText)
		flag.PrintDefaults()
	}

	flag.BoolVarP(&version, "version", "V", false, "display version information and exit")
	flag.BoolVarP(&debug, "debug", "D", false, "enable debug logging")

	flag.StringVarP(&bind, "bind", "b", "0.0.0.0:8000", "[interface]:port to bind to")
	flag.StringVarP(&dir, "dir", "d", "./data", "path to store data in")
}

func main() {
	flag.Parse()

	if debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	if version {
		fmt.Printf("fbox version %s", FullVersion())
		os.Exit(0)
	}

	dir := os.ExpandEnv(dir)
	if err := os.MkdirAll(dir, 0700); err != nil {
		log.WithError(err).Fatalf("error creating directory %s", dir)
	}
	dir = os.ExpandEnv(dir)
	storage := store.NewDiskStore(dir)
	log.Infof("using %s for storage", storage)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var logger *log.Entry
		status, body := func() (int, []byte) {
			hkey := r.URL.Path[1:]
			key, err := hex.DecodeString(hkey)
			if err != nil {
				return http.StatusBadRequest, []byte(fmt.Sprintf("%q: not a valid path, expecting hex key only", r.URL.Path))
			}
			logger = log.WithFields(log.Fields{
				"op":  r.Method,
				"key": hkey,
			})
			switch r.Method {
			case http.MethodGet:
				value, err := storage.Get(key)
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
				if err := storage.Put(key, value); err != nil {
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
	})

	app := logger.New(logger.Options{
		Prefix:               "fbox",
		RemoteAddressHeaders: []string{"X-Forwarded-For"},
	}).Handler(http.DefaultServeMux)

	log.Infof("fbox %s listening on %s", FullVersion(), bind)
	if err := http.ListenAndServe(bind, app); err != nil {
		log.WithField("err", err).Fatal("Could not listen and serve")
	}
}
