package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/go-sockaddr/template"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"github.com/unrolled/logger"

	"github.com/prologic/bitcask"
	"github.com/prologic/fbox/blob"
	"github.com/prologic/fbox/store"
	"xojoc.pw/useragent"
)

var (
	debug   bool
	version bool

	bind             string
	master           string
	advertiseAddress string
	dataShards       int
	parityShards     int
	dir              string

	//go:embed ui
	uiFS embed.FS
)

const helpText = `
fbox is a simple distributed file system...

Valid commands:
 - cat <name> -- Downloads the given file given by <name> to stdout
 - put <name> -- Upload the given file given by <name>

Valid options:
`

const (
	maxPathLength = 256 // TODO: Should this be configurable?
)

func init() {
	baseProg := filepath.Base(os.Args[0])
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [command [arguments]]\n", baseProg)
		fmt.Fprint(os.Stderr, helpText)
		flag.PrintDefaults()
	}

	flag.BoolVarP(&version, "version", "V", false, "display version information and exit")
	flag.BoolVarP(&debug, "debug", "D", false, "enable debug logging")

	flag.StringVarP(&bind, "bind", "b", "0.0.0.0:8000", "[interface]:port to bind to")
	flag.StringVarP(&master, "master", "m", "", "address:port of master")
	flag.StringVarP(&advertiseAddress, "advertise-addr", "a", "", "[interface]:port to advertise")
	flag.StringVarP(&dir, "dir", "d", "./data", "path to store data in")
	flag.IntVarP(&dataShards, "data-shards", "s", 3, "no. of data shards")
	flag.IntVarP(&parityShards, "parity-shards", "p", 1, "no. of parity shards")
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

	log.SetReportCaller(true)

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

	switch strings.ToLower(flag.Arg(0)) {
	case "cat":
		status, err := cmdCat(mAddr, flag.Args()[1:])
		if err != nil {
			log.WithError(err).Error("error reading file")
		}
		os.Exit(status)
	case "put":
		status, err := cmdPut(mAddr, flag.Args()[1:])
		if err != nil {
			log.WithError(err).Error("error writing file")
		}
		os.Exit(status)
	}

	dir := os.ExpandEnv(dir)
	if err := os.MkdirAll(dir, 0700); err != nil {
		log.WithError(err).Fatalf("error creating directory %s", dir)
	}
	dir = os.ExpandEnv(dir)
	storage := store.NewDiskStore(dir)
	log.Infof("using %s for storage", storage)

	http.Handle(
		"/blob/",
		http.StripPrefix(
			"/blob/",
			blob.NewHandler(store.NewDiskStore(dir)),
		),
	)

	if mAddr == "" {
		http.HandleFunc("/join", joinHandler)
		http.HandleFunc("/nodes", nodesHandler)
		http.HandleFunc("/files", filesHandler)
		http.Handle("/metadata/", http.StripPrefix("/metadata/", http.HandlerFunc(metadataHandler)))
		http.Handle("/upload/", http.StripPrefix("/upload/", http.HandlerFunc(uploadHandler)))
		http.Handle("/download/", http.StripPrefix("/download/", http.HandlerFunc(downloadHandler)))

		var err error

		dbPath := filepath.Join(dir, "meta.db")
		db, err = bitcask.Open(
			dbPath,
			// TODO: Make this configurable?
			// XXX: Or figure out a better way to store paths?
			bitcask.WithMaxKeySize(maxPathLength),
		)

		if err != nil {
			log.WithError(err).Fatalf("error opening metdata db %s", dir)
		}
		log.Infof("storing metdata at %s using bitcask", dbPath)

		// Serve up the UI
		if debug {
			http.Handle("/ui/", http.StripPrefix("/ui/", http.FileServer(http.Dir("./ui/"))))
		} else {
			http.Handle("/ui/", http.FileServer(http.FS(uiFS)))
		}
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			ua := useragent.Parse(r.UserAgent())
			if ua != nil && ua.Type == useragent.Browser {
				http.Redirect(w, r, "/ui/", http.StatusFound)
			}
		})

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
		go func() {
			time.Sleep(time.Second * 3)
			if err := joinNode(aAddr, mAddr); err != nil {
				log.WithError(err).Warnf("error joining node %s (retrying in 1s...)", mAddr)
			}
			log.Infof("successfully joined master node %s", master)
			return
		}()
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
