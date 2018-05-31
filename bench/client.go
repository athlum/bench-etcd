package benchEtcd

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"strings"
	"sync"
	"time"
)

type latency struct {
	lock  *sync.RWMutex
	value time.Duration
	count int
}

type manage struct {
	conns     int
	clients   int
	endpoints string
	keySet    *keySet
	cache     *sync.Map
	latency   *latency
}

type keySet struct {
	key       string
	keys      int
	valueSize int
}

func (k *keySet) keylist() []string {
	ks := make([]string, k.keys)
	for i := 0; i < k.keys; i += 1 {
		ks[i] = fmt.Sprintf("/%v/%v", k.key, uuid())
	}
	return ks
}

func (l *latency) submit(d time.Duration) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.count += 1
	l.value += d
}

func (l *latency) avg() float64 {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return float64(l.value) / float64(l.count*int(time.Millisecond))
}

func (m *manage) run(l *loop) {
	kl := m.keySet.keylist()
	m.cacheInit(kl, m.keySet.valueSize, m.ec())
	fmt.Println("Init Done")
	ch := l.totalC(kl)
	for i := 0; i < m.conns; i += 1 {
		go m.cli(l, ch)
	}
}

func (m *manage) cluster() []string {
	return strings.Split(m.endpoints, ",")
}

func (m *manage) ec() *clientv3.Client {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   m.cluster(),
		DialTimeout: time.Second,
	})
	if err != nil {
		panic(err)
	}
	return cli
}

func (m *manage) cli(l *loop, ch <-chan string) {
	cli := m.ec()
	defer cli.Close()

	for i := 0; i < m.clients/m.conns; i += 1 {
		c := client{
			cli:    cli,
			keySet: m.keySet,
		}
		go c.run(ch, l, m)
	}
	l.wait()
}

type client struct {
	cli    *clientv3.Client
	keySet *keySet
}

func (c *client) run(ch <-chan string, l *loop, m *manage) {
	for k := range ch {
		if kk, ok := m.cache.Load(k); ok {
			kk.(*key).newValueWatch(c.keySet.valueSize, m, func(value string) {
				_, err := c.cli.Put(context.Background(), k, value)
				if err != nil {
					panic(err)
				}
			})
			l.done()
		} else {
			panic(fmt.Sprintf("keyname %v cache not found.", k))
		}
	}
}
