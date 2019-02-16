/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"strings"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/xyproto/simpleredis"

	newrelic "github.com/newrelic/go-agent"
	log "github.com/sirupsen/logrus"
)

var (
	nrapp      newrelic.Application
	masterPool *simpleredis.ConnectionPool
	slavePool  *simpleredis.ConnectionPool
)

type config struct {
	serverAddr   string
	nrLicenseKey string
	redisMaster  string
	redisSlave   string
	debug        bool
}

func ListRangeHandler(rw http.ResponseWriter, req *http.Request) {
	key := mux.Vars(req)["key"]
	list := simpleredis.NewList(slavePool, key)
	members := HandleError(list.GetAll()).([]string)
	membersJSON := HandleError(json.MarshalIndent(members, "", "  ")).([]byte)
	rw.Write(membersJSON)
}

func ListPushHandler(rw http.ResponseWriter, req *http.Request) {
	key := mux.Vars(req)["key"]
	value := mux.Vars(req)["value"]
	list := simpleredis.NewList(masterPool, key)
	HandleError(nil, list.Add(value))
	ListRangeHandler(rw, req)
}

func InfoHandler(rw http.ResponseWriter, req *http.Request) {
	info := HandleError(masterPool.Get(0).Do("INFO")).([]byte)
	rw.Write(info)
}

func EnvHandler(rw http.ResponseWriter, req *http.Request) {
	environment := make(map[string]string)
	for _, item := range os.Environ() {
		splits := strings.Split(item, "=")
		key := splits[0]
		val := strings.Join(splits[1:], "=")
		environment[key] = val
	}

	envJSON := HandleError(json.MarshalIndent(environment, "", "  ")).([]byte)
	rw.Write(envJSON)
}

func HandleError(result interface{}, err error) (r interface{}) {
	if err != nil {
		log.Error(err)
	}
	return result
}

// NrMiddleware is a quick and dirty negroni middleware for transactions
//  HACK! uses global nrapp....
// use:
//   n := negroni.Classic()
//   n.Use(negroni.HandlerFunc(NrMiddleware)) // <= Important part here
//   n.UseHandler(r)  // Creating handler outside of scope
//   n.Run(cfg.serverAddr)
func NrMiddleware(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if nrapp != nil {
		trname := "Unknown"

		// Probably bad, as it contains the IDs...
		if r != nil && r.URL != nil {
			trname = r.URL.Path
		}

		trx := nrapp.StartTransaction(trname, rw, r)
		defer trx.End()
	}

	next(rw, r)
}

func main() {
	var err error
	var cfg config

	// Parse cli options: https://golang.org/pkg/flag/
	flag.StringVar(&cfg.serverAddr, "server", ":3000", "Address for server to listen on")
	flag.StringVar(&cfg.nrLicenseKey, "nrkey", "", "New Relic License Key")
	flag.StringVar(&cfg.redisMaster, "redis", "localhost:6379", "Redis Master")
	flag.StringVar(&cfg.redisSlave, "redisslave", "localhost:6379", "Redis Slave")
	flag.BoolVar(&cfg.debug, "debug", false, "Debug mode")
	flag.Parse()

	if cfg.debug {
		log.SetLevel(log.DebugLevel) //Verbose
	}

	config := newrelic.NewConfig("GuestBook-Go-NewRelic", cfg.nrLicenseKey)

	if cfg.debug {
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

	n := negroni.Classic()
	n.Use(negroni.HandlerFunc(NrMiddleware))
	n.UseHandler(r)
	n.Run(cfg.serverAddr)
}
