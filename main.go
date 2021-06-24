package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/go-sockaddr/template"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"github.com/unrolled/logger"

	"github.com/prologic/fbox/blob"
	"github.com/prologic/fbox/store"
)

var (
	debug   bool
	version bool

	bind             string
	master           string
	advertiseAddress string
	dir              string
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
	flag.StringVarP(&master, "master", "m", "", "address:port of master")
	flag.StringVarP(&advertiseAddress, "advertise-addr", "a", "", "[interface]:port to advertise")
	flag.StringVarP(&dir, "dir", "d", "./data", "path to store data in")
}

func mustParseAddress(addr string) string {
	r, err := template.Parse(addr)
	if err != nil {
		log.WithError(err).Fatalf("error parsing addr %s", addr)
	}
	return r
}

func joinNode(aAddr, mAddr string) error {
	data, _ := json.Marshal(map[string]string{"addr": aAddr})
	buf := bytes.NewReader(data)
	res, err := http.Post(mAddr+"/join", "application/json", buf)
	if err != nil {
		log.WithError(err).Error("error making join request")
		return fmt.Errorf("error making join request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.WithError(err).Error("non-200 recieved from join request")
		return fmt.Errorf("non-200 recieved from join request: %s", res.Status)
	}

	return nil
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

	bAddr := mustParseAddress(bind)
	mAddr := mustParseAddress(master)
	aAddr := mustParseAddress(advertiseAddress)

	dir := os.ExpandEnv(dir)
	if err := os.MkdirAll(dir, 0700); err != nil {
		log.WithError(err).Fatalf("error creating directory %s", dir)
	}
	dir = os.ExpandEnv(dir)
	storage := store.NewDiskStore(dir)
	log.Infof("using %s for storage", storage)

	http.Handle("/blobs", blob.NewHandler(store.NewDiskStore(dir)))

	if mAddr == "" {
		http.HandleFunc("/join", joinHandler)
		http.HandleFunc("/nodes", nodesHandler)
		http.HandleFunc("/upload", uploadHandler)
		http.HandleFunc("/download", downloadHandler)

		// Join ourself
		go func() {
			time.Sleep(time.Second * 3)
			if err := joinNode(aAddr, fmt.Sprintf("http://%s", aAddr)); err != nil {
				log.WithError(err).Fatalf("error joining node %s", mAddr)
			}
			log.Infof("successfully joined master node %s", master)
		}()
	} else {
		// Join an existing node
		if err := joinNode(aAddr, mAddr); err != nil {
			log.WithError(err).Fatalf("error joining node %s", mAddr)
		}
		log.Infof("successfully joined master node %s", master)
	}

	app := logger.New(logger.Options{
		Prefix:               "fbox",
		RemoteAddressHeaders: []string{"X-Forwarded-For"},
	}).Handler(http.DefaultServeMux)

	log.Infof("fbox %s listening on %s", FullVersion(), bAddr)
	if err := http.ListenAndServe(bind, app); err != nil {
		log.WithField("err", err).Fatal("Could not listen and serve")
	}
}
