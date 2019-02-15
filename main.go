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
    masterPool *simpleredis.ConnectionPool
    slavePool  *simpleredis.ConnectionPool
)

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
        panic(err)
    }
    return result
}

func main() {
    config := newrelic.NewConfig("GuestBook-Go-NewRelic", "New-Relic-Key")
    app, err := newrelic.NewApplication(config)
    if err != nil {
        log.Fatalf("Unable to create NR connection: %v", err)
    }

    masterPool = simpleredis.NewConnectionPoolHost("redis-master:6379")
    defer masterPool.Close()
    slavePool = simpleredis.NewConnectionPoolHost("redis-slave:6379")
    defer slavePool.Close()

    r := mux.NewRouter()
    r.HandleFunc(newrelic.WrapHandleFunc(app, "/lrange/{key}", ListRangeHandler)).Methods("GET")
    r.HandleFunc(newrelic.WrapHandleFunc(app, "/rpush/{key}/{value}", ListPushHandler)).Methods("GET")
    r.HandleFunc(newrelic.WrapHandleFunc(app, "/info", InfoHandler)).Methods("GET")
    r.HandleFunc(newrelic.WrapHandleFunc(app, "/env", EnvHandler)).Methods("GET")

    n := negroni.Classic()
    n.UseHandler(r)
    n.Run(":3000")
}
