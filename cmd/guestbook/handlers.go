package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/xyproto/simpleredis"
)

// ListRangeHandler ...
func ListRangeHandler(rw http.ResponseWriter, req *http.Request) {
	key := mux.Vars(req)["key"]
	list := simpleredis.NewList(slavePool, key)
	members := logIfError(list.GetAll()).([]string)
	membersJSON := logIfError(json.MarshalIndent(members, "", "  ")).([]byte)

	logIfError(rw.Write(membersJSON))
}

// ListPushHandler ...
func ListPushHandler(rw http.ResponseWriter, req *http.Request) {
	key := mux.Vars(req)["key"]
	value := mux.Vars(req)["value"]
	list := simpleredis.NewList(masterPool, key)

	logIfError(nil, list.Add(value))

	ListRangeHandler(rw, req)
}

// InfoHandler ...
func InfoHandler(rw http.ResponseWriter, req *http.Request) {
	info := logIfError(masterPool.Get(0).Do("INFO")).([]byte)

	logIfError(rw.Write(info))
}

// EnvHandler returns our environment as a dataset
func EnvHandler(rw http.ResponseWriter, req *http.Request) {
	environment := make(map[string]string)
	for _, item := range os.Environ() {
		splits := strings.Split(item, "=")
		key := splits[0]
		val := strings.Join(splits[1:], "=")
		environment[key] = val
	}

	envJSON := logIfError(json.MarshalIndent(environment, "", "  ")).([]byte)

	logIfError(rw.Write(envJSON))
}
