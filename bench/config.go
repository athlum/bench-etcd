package benchEtcd

import (
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
