package benchEtcd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var Bench = &cobra.Command{
	Use:   "bench-etcd",
	Short: "Etcd bench mark tool",
	Long:  "Etcd bench mark tool",
	Run: func(cmd *cobra.Command, args []string) {
		data, err := json.Marshal(config)
		if err != nil {
			panic(err)
		}
		fmt.Printf("config: %v", string(data))
		now := time.Now()
		config.manage.run(config.loop)
		config.loop.wait()
		fmt.Println("end", "time:", float64(time.Now().Sub(now).Nanoseconds())/float64(time.Microsecond))
	},
}

func init() {
	config = newCfg()
	Bench.Flags().IntVarP(&config.loop.total, "total", "t", 10000, "total put requests")
	Bench.Flags().IntVarP(&config.manage.clients, "clients", "l", 1, "total clients")
	Bench.Flags().IntVarP(&config.manage.conns, "conns", "c", 1, "total conns")
	Bench.Flags().StringVarP(&config.manage.endpoints, "endpoints", "e", "", "endpoints")
	Bench.Flags().IntVarP(&config.manage.keySet.keys, "keys", "s", 10000, "key number")
	Bench.Flags().IntVarP(&config.manage.keySet.valueSize, "valueSize", "v", 256, "value length")
	Bench.Flags().StringVarP(&config.manage.keySet.key, "key", "k", "key", "key name")
	Bench.MarkFlagRequired("endpoints")
	Bench.ParseFlags(os.Args)
}