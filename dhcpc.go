package ippool

import (
	"net"
	"sync"
)

type Dhcpc struct {
	genIP4 *GenIP4
	pool4  map[uint32]struct{}
	mu4    sync.Mutex

	Network *net.IPNet
}

func New(network string) (*Dhcpc, error) {
	genIP4, err := NewGenIP4(network)
	if err != nil {
		return nil, err
	}

	_, ipNet, err := net.ParseCIDR(network)
	if err != nil {
		return nil, err
	}

	return &Dhcpc{
		genIP4:  genIP4,
		pool4:   map[uint32]struct{}{},
		mu4:     sync.Mutex{},
		Network: ipNet,
	}, nil
}

// IPv4 returns the next free IPv4 or nil if the pool is exhausted.
func (d *Dhcpc) IPv4() net.IP {
	octet0, octet1, octet2, octet3 := d.genIP4.NextIP()

	d.mu4.Lock()
	defer d.mu4.Unlock()
	id := uint32(octet0)<<24 | uint32(octet1)<<16 | uint32(octet2)<<8 | uint32(octet3)
	// "fast path"
	if _, ok := d.pool4[id]; !ok {
		d.pool4[id] = struct{}{}
		return net.IP{octet0, octet1, octet2, octet3}
	}

	// here it can get messy if the pool is near to exhaustion
	for {
		o0, o1, o2, o3 := d.genIP4.NextIP()
		id := uint32(o0)<<24 | uint32(o1)<<16 | uint32(o2)<<8 | uint32(o3)
		if _, ok := d.pool4[id]; !ok {
			d.pool4[id] = struct{}{}
			return net.IP{o0, o1, o2, o3}
		}

		if octet0 == o0 && octet1 == o1 && octet2 == o2 && octet3 == o3 {
			// did a complete spin, no free IPv4 addresses were found
			return nil
		}
	}
}

func (d *Dhcpc) Release4(ip net.IP) {
	d.mu4.Lock()
	defer d.mu4.Unlock()

	id := uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
	delete(d.pool4, id)
}
