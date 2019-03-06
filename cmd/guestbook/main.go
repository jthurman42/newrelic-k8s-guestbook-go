package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"github.com/xyproto/simpleredis"

	newrelic "github.com/newrelic/go-agent"
	nrgorilla "github.com/newrelic/go-agent/_integrations/nrgorilla/v1"
	log "github.com/sirupsen/logrus"
)

const (
	appName = "GuestBook-Go-NewRelic"
)

var (
	version    = "unknown"
	masterPool *simpleredis.ConnectionPool
	slavePool  *simpleredis.ConnectionPool
)

type config struct {
	serverAddr     string
	nrLicenseKey   string
	assetDirectory string
	redisMaster    string
	redisSlave     string
	debug          bool
}

func main() {
	var nrapp newrelic.Application
	var err error
	var cfg config

	log.Infof("Starting %s version %s", appName, version)

	// Parse cli options: https://golang.org/pkg/flag/
	flag.StringVar(&cfg.serverAddr, "server", ":3000", "Address for server to listen on")
	flag.StringVar(&cfg.nrLicenseKey, "nrkey", "", "New Relic License Key")
	flag.StringVar(&cfg.assetDirectory, "assets", "./public", "Static asset directory")
	flag.StringVar(&cfg.redisMaster, "redis", "localhost:6379", "Redis Master")
	flag.StringVar(&cfg.redisSlave, "redisslave", "localhost:6379", "Redis Slave")
	flag.BoolVar(&cfg.debug, "debug", false, "Debug mode")
	flag.Parse()

	// Create NR Agent Configuration
	config := newrelic.NewConfig(appName, cfg.nrLicenseKey)

	// Enable Distributed Tracing
	config.CrossApplicationTracer.Enabled = false
	config.DistributedTracer.Enabled = true

	if cfg.debug {
		log.SetLevel(log.DebugLevel) //Verbose
		config.Logger = newrelic.NewDebugLogger(os.Stdout)
	}

	nrapp, err = newrelic.NewApplication(config)
	if err != nil {
		log.Fatalf("Unable to create NR connection: %v", err)
	}

	masterPool = simpleredis.NewConnectionPoolHost(cfg.redisMaster)
	defer masterPool.Close()
	slavePool = simpleredis.NewConnectionPoolHost(cfg.redisSlave)
	defer slavePool.Close()

	r := mux.NewRouter()
	r.HandleFunc("/lrange/{key}", ListRangeHandler).Methods("GET")
	r.HandleFunc("/rpush/{key}/{value}", ListPushHandler).Methods("GET")
	r.HandleFunc("/info", InfoHandler).Methods("GET")
	r.HandleFunc("/env", EnvHandler).Methods("GET")

	// Should be last
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(cfg.assetDirectory)))

	logIfError(nil, http.ListenAndServe(cfg.serverAddr, nrgorilla.InstrumentRoutes(r, nrapp)))
}
