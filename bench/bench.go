package benchEtcd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var Bench = &cobra.Command{
	Use:   "bench-etcd",
	Short: "Etcd bench mark tool",
	Long:  "Etcd bench mark tool",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("config: %v\n", config.JSON())
		config.manage.run(config.loop)
		config.loop.wait()
		fmt.Println("timeout:", config.manage.timeout.val())
		fmt.Println("watch failed:", config.manage.watchFailed.val())
		fmt.Println("avg latency:", config.manage.latency.avg())
		fmt.Println("max latency:", config.manage.latency.max())
		m := config.manage.latency.distribution()
		fmt.Println("latency distribution:")
		for c, v := range m {
			fmt.Printf("<%vs: %v", c+1, v)
		}
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
