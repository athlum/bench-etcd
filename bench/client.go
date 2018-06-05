package benchEtcd

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"strings"
	"sync"
	"time"
)

type counter struct {
	lock  *sync.RWMutex
	value int
}

func (c *counter) echo() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.value += 1
}

func (c *counter) val() int {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.value
}

type latency struct {
	lock    *sync.RWMutex
	value   time.Duration
	count   int
	longest time.Duration
	m       map[int]int
}

type manage struct {
	conns       int
	clients     int
	endpoints   string
	keySet      *keySet
	cache       *sync.Map
	latency     *latency
	timeout     *counter
	watchFailed *counter
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
	if d > l.longest {
		l.longest = d
	}
	l.m[int(d/time.Second)] += 1
}

func (l *latency) avg() time.Duration {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return time.Duration(int(l.value) / l.count)
}

func (l *latency) progress() int {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return l.count
}

func (l *latency) notice() {
	t := time.NewTicker(time.Second * 300)
	defer t.Stop()
	for {
		<-t.C
		fmt.Println("progress:", l.progress())
	}
}

func (l *latency) max() time.Duration {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return l.longest
}

func (l *latency) distribution() map[int]int {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return l.m
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
		DialTimeout: time.Second * 10,
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
				for {
					ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
					_, err := c.cli.Put(ctx, k, value)
					if err != nil {
						c.cli = m.ec()
						m.timeout.echo()
						fmt.Println("put failed", err)
					} else {
						return
					}
				}
			})
			l.done()
		} else {
			panic(fmt.Sprintf("keyname %v cache not found.", k))
		}
	}
}
