package lb

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

var (
	ErrNoHostsToBalance = errors.New("no hosts to balance")
)

type RoundRobin struct {
	hosts   []string
	current int
	rand    rand.Source
	locker  sync.Locker
}

func NewRoundRobin(hosts []string) *RoundRobin {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	return &RoundRobin{
		hosts:   hosts,
		rand:    rand,
		current: rand.Intn(len(hosts)),
		locker:  &sync.Mutex{},
	}
}

func (b *RoundRobin) Upgrade(hosts []string) {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.hosts = hosts
	b.current = rand.Intn(len(hosts))
}

func (b *RoundRobin) Size() int {
	b.locker.Lock()
	defer b.locker.Unlock()

	return len(b.hosts)
}

func (b *RoundRobin) Next() (string, error) {
	b.locker.Lock()
	defer b.locker.Unlock()

	if len(b.hosts) == 0 {
		return "", ErrNoHostsToBalance
	}
	if len(b.hosts) == 1 {
		return b.hosts[0], nil
	}
	host := b.hosts[b.current]
	b.current = (b.current + 1) % len(b.hosts)

	return host, nil
}
