package ippool

import (
	"errors"
	"strconv"
	"strings"
	"sync/atomic"
)

const (
	mask24 = uint32(0xFF << 24)
	mask16 = uint32(0xFF << 16)
	mask8  = uint32(0xFF << 8)
	mask0  = uint32(0xFF)
)

// GenIP4 is a sequential IPv4 generator.
type GenIP4 struct {
	maskHostBits uint32
	counter      uint32
	network      uint32
	broadcast    uint32
}

// NewGenIP4 creates a new IPv4 address's generator for the given CIDR network, returning an
// error if it's not in the right format or is invalid.
func NewGenIP4(network string) (*GenIP4, error) {
	parts := strings.Split(network, "/")
	if len(parts) != 2 {
		return nil, errors.New("invalid CIDR network")
	}

	maskbits, err := strconv.ParseUint(parts[1], 10, 6)
	if err != nil {
		return nil, errors.New("invalid mask bits")
	}
	if maskbits > 32 {
		return nil, errors.New("invalid mask bits")
	}

	octetsRaw := strings.Split(parts[0], ".")
	if len(octetsRaw) != 4 {
		return nil, errors.New("invalid ip format")
	}

	octets := []uint{}
	for _, oct := range octetsRaw {
		octet, err := strconv.ParseUint(oct, 10, 8)
		if err != nil {
			return nil, errors.New("invalid ip octet")
		}
		octets = append(octets, uint(octet))
	}

	// invert mask
	maskHostBits := uint32(0xFFFFFFFF >> maskbits)
	ipNetwork := uint32(octets[0]<<24|octets[1]<<16|octets[2]<<8|octets[3]<<0) & ^maskHostBits
	gen := GenIP4{
		maskHostBits: maskHostBits,
		network:      ipNetwork,
		broadcast:    maskHostBits | ipNetwork,
	}

	return &gen, nil
}

func (p *GenIP4) NextIP() (byte, byte, byte, byte) {
	counter := p.counter
	for !atomic.CompareAndSwapUint32(&p.counter, counter, counter+1) {
		counter = p.counter
	}

	hostHalf := counter & p.maskHostBits

	octet0 := byte((p.network | (hostHalf & mask24)) >> 24)
	octet1 := byte((p.network | (hostHalf & mask16)) >> 16)
	octet2 := byte((p.network | (hostHalf & mask8)) >> 8)
	octet3 := byte((p.network | (hostHalf & mask0)) >> 0)

	return octet0, octet1, octet2, octet3
}
