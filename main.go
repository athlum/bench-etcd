package main

import (
	"github.com/athlum/bench-etcd/bench"
	"net/http"
	_ "net/http/pprof"
)

func main() {
	go func() {
		http.ListenAndServe("0.0.0.0:10030", nil)
	}()
	if err := benchEtcd.Bench.Execute(); err != nil {
		panic(err)
	}
}
