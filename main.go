package main

import (
	"github.com/athlum/bench-etcd/bench"
)

func main() {
	if err := benchEtcd.Bench.Execute(); err != nil {
		panic(err)
	}
}
