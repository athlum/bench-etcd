package benchEtcd

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/satori/go.uuid"
	"strings"
	"time"
)

type manage struct {
	conns     int
	clients   int
	endpoints string
	keySet    *keySet
}

type keySet struct {
	key       string
	keys      int
	valueSize int
}

func (k *keySet) keylist() []string {
	ks := make([]string, k.keys)
	for i := 0; i < k.keys; i += 1 {
		u, err := uuid.NewV4()
		if err != nil {
			panic(err)
		}
		ks[i] = fmt.Sprintf("/%v/%v", k.key, u.String())
	}
	return ks
}

func (m *manage) run(l *loop) {
	for i := 0; i < m.conns; i += 1 {
		go m.cli(l)
	}
}

func (m *manage) cluster() []string {
	return strings.Split(m.endpoints, ",")
}

func (m *manage) cli(l *loop) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   m.cluster(),
		DialTimeout: time.Second,
	})
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	ch := l.totalC(m.keySet.keylist())
	for i := 0; i < m.clients; i += 1 {
		c := client{
			cli:    cli,
			keySet: m.keySet,
		}
		go c.run(ch, l)
	}
	l.wait()
}

type client struct {
	cli    *clientv3.Client
	keySet *keySet
}

func (c *client) run(ch <-chan string, l *loop) {
	for k := range ch {
		_, err := c.cli.Put(context.Background(), k, strings.Repeat("v", c.keySet.valueSize))
		if err != nil {
			panic(err)
		}
		l.done()
	}
}
