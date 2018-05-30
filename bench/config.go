package benchEtcd

import (
	"encoding/json"
	"sync"
)

var config *cfg

type loop struct {
	total int
	wg    *sync.WaitGroup
}

func (l *loop) totalC(kl []string) <-chan string {
	c := make(chan string, l.total)
	for i := 1; i <= l.total; i += 1 {
		l.wg.Add(i)
		c <- kl[(i-1)%len(kl)]
	}
	return c
}

func (l *loop) wait() {
	l.wg.Wait()
}

func (l *loop) done() {
	l.wg.Done()
}

type cfg struct {
	loop   *loop
	manage *manage
}

func (c *cfg) JSON() string {
	data, err := json.Marshal(&struct {
		Total       int
		Clients     int
		Connections int
		Endpoints   string
		KeyPrefix   string
		KeyNumber   int
		ValueSize   int
	}{
		Total:       c.loop.total,
		Clients:     c.manage.clients,
		Connections: c.manage.conns,
		Endpoints:   c.manage.endpoints,
		KeyPrefix:   c.manage.keySet.key,
		KeyNumber:   c.manage.keySet.keys,
		ValueSize:   c.manage.keySet.valueSize,
	})
	if err != nil {
		panic(err)
	}
	return string(data)
}

func newCfg() *cfg {
	return &cfg{
		loop: &loop{
			wg: &sync.WaitGroup{},
		},
		manage: &manage{
			keySet: &keySet{},
		},
	}
}
