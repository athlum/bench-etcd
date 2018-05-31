package benchEtcd

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"strings"
	"sync"
	"time"
)

func (m *manage) cacheInit(keylist []string, valueSize int, cli *clientv3.Client) {
	for _, k := range keylist {
		kk := &key{
			lock:    &sync.Mutex{},
			keyName: k,
		}
		kk.newValue(valueSize, cli)
	}
	defer cli.Close()
}

type key struct {
	lock    *sync.Mutex
	keyName string
	value   string
	time    time.Time
}

func (k *key) _newValue(valueSize int) {
	u := uuid()
	k.value = fmt.Sprintf("%v%v", strings.Repeat("v", valueSize-len(u)), u)
}

func (k *key) newValue(valueSize int, cli *clientv3.Client) {
	k.lock.Lock()
	defer k.lock.Unlock()

	k._newValue(valueSize)
	_, err := cli.Put(context.Background(), k.keyName, k.value)
	if err != nil {
		panic(err)
	}

	k.time = time.Now()
}

func (k *key) newValueWatch(valueSize int, m *manage, fn func(string)) {
	k.lock.Lock()
	defer k.lock.Unlock()

	cli := m.ec()
	defer cli.Close()

	k._newValue(valueSize)
	wc := cli.Watch(context.Background(), k.keyName)
	for e := range wc {
		if err := e.Err(); err != nil {
			panic(err)
		}
		for _, ev := range e.Events {
			if k.value == string(ev.Kv.Value) {
				m.latency.submit(time.Now().Sub(k.time))
				return
			}
		}
	}

	go fn(k.value)
}