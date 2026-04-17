// Package lb provides load balancing functionality with a Round Robin algorithm.
//
// The package offers a thread-safe RoundRobin balancer that distributes requests
// across a list of hosts in a cyclic manner. It supports dynamic host list updates
// and is safe for concurrent use by multiple goroutines.
//
// Basic usage:
//
//	hostList := []string{"localhost:8080", "localhost:8081"}
//	balancer := lb.NewRoundRobin(hostList)
//	host, err := balancer.Next()
package lb

import (
	"errors"
	"math/rand"
	"sync"
)

var (
	// ErrNoHostsToBalance is returned when no hosts are available for balancing.
	ErrNoHostsToBalance = errors.New("no hosts to balance")
)

// RoundRobin implements a Round Robin load balancing algorithm.
// It distributes requests across a list of hosts in a cyclic order.
// The RoundRobin balancer is safe for concurrent use by multiple goroutines.
type RoundRobin struct {
	hosts   []string
	current int
	locker  sync.Locker
}

// NewRoundRobin creates a new RoundRobin balancer with the provided list of hosts.
// The initial position is set to a random index if hosts are provided.
// It returns a pointer to the newly created RoundRobin instance.
func NewRoundRobin(hosts []string) *RoundRobin {
	current := 0
	if len(hosts) > 0 {
		current = rand.Intn(len(hosts))
	}
	return &RoundRobin{
		hosts:   hosts,
		current: current,
		locker:  &sync.Mutex{},
	}
}

// Upgrade updates the list of hosts to balance.
// It resets the current position to a random index if hosts are provided.
// This method is safe for concurrent use.
func (b *RoundRobin) Upgrade(hosts []string) {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.hosts = hosts
	current := 0
	if len(hosts) > 0 {
		current = rand.Intn(len(hosts))
	}
	b.current = current
}

// Size returns the current number of hosts in the balancer.
// This method is safe for concurrent use.
func (b *RoundRobin) Size() int {
	b.locker.Lock()
	defer b.locker.Unlock()

	return len(b.hosts)
}

// Next returns the next host from the list using the Round Robin algorithm.
// It cycles through the hosts in order, returning ErrNoHostsToBalance if the list is empty.
// This method is safe for concurrent use.
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
