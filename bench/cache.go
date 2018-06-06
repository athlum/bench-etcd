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
	wg := &sync.WaitGroup{}
	wg.Add(len(keylist))
	for _, k := range keylist {
		go func(wg *sync.WaitGroup, k string) {
			defer wg.Done()
			kk := &key{
				lock:    &sync.Mutex{},
				keyName: k,
			}
			kk.newValue(valueSize, cli)
			m.cache.Store(k, kk)
		}(wg, k)
	}
	wg.Wait()
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
	ctx, _ := context.WithTimeout(context.Background(), time.Second*20)
	wc := cli.Watch(ctx, k.keyName)
	fn(k.value)
	k.time = time.Now()
	t := time.NewTicker(time.Second * 60)
	defer t.Stop()
	for {
		select {
		case e := <-wc:
			if err := e.Err(); err != nil {
				fmt.Println("watch Failed:", err.Error())
				m.watchFailed.echo()
				return
			}
			for _, ev := range e.Events {
				if k.value == string(ev.Kv.Value) {
					m.latency.submit(time.Now().Sub(k.time))
					return
				}
			}
		case <-t.C:
			m.watchFailed.echo()
			return
		}
	}
}
