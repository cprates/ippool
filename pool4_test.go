package ippool

import (
	"errors"
	"net"
	"sync/atomic"
	"testing"
)

func TestPool4(t *testing.T) {
	testsSet := []struct {
		description      string
		network          string
		wantErr          error
		wantNetwork      uint32
		wantMaskHostBits uint32
	}{
		{
			description: "invalid_network_format",
			network:     "10.0.0.1/4/5",
			wantErr:     errors.New("invalid CIDR network"),
		},
		{
			description: "invalid_network_bits_1",
			network:     "10.0.0.1/33",
			wantErr:     errors.New("invalid mask bits"),
		},
		{
			description: "invalid_network_bits_2",
			network:     "10.0.0.1/abc",
			wantErr:     errors.New("invalid mask bits"),
		},
		{
			description: "invalid_network_bits_3",
			network:     "10.0.0.1/",
			wantErr:     errors.New("invalid mask bits"),
		},
		{
			description: "invalid_ip_format_1",
			network:     "10.0.01/24",
			wantErr:     errors.New("invalid ip format"),
		},
		{
			description: "invalid_ip_format_2",
			network:     "10.0.0.1.4/24",
			wantErr:     errors.New("invalid ip format"),
		},
		{
			description: "invalid_ip_with_char_in_octet",
			network:     "a.b.c.d/24",
			wantErr:     errors.New("invalid ip octet"),
		},
		{
			description: "ip_with_char_in_octet",
			network:     "a.b.c.d/24",
			wantErr:     errors.New("invalid ip octet"),
		},
		{
			description: "ip_with_octet_over_255",
			network:     "10.256.0.1/24",
			wantErr:     errors.New("invalid ip octet"),
		},
		{
			description:      "ip_octet_boundaries_0",
			network:          "10.0.0.1/24",
			wantNetwork:      167772160,
			wantMaskHostBits: 255,
		},
		{
			description:      "ip_octet_boundaries_255",
			network:          "10.0.255.1/8",
			wantNetwork:      167772160,
			wantMaskHostBits: 16777215,
		},
		{
			description:      "match_global4counter",
			network:          "153.153.153.153/28",
			wantNetwork:      2576980368,
			wantMaskHostBits: 15,
		},
	}

	for _, test := range testsSet {
		t.Run(test.description,
			func(t *testing.T) {
				pool, err := NewPool4(test.network)
				if test.wantErr != nil {
					if err == nil {
						t.Errorf("want error %q but got nil", test.wantErr)
					}

					if test.wantErr.Error() != err.Error() {
						t.Errorf("want error %q got %q", test.wantErr, err)
					}

					return
				}
				if err != nil {
					t.Errorf("doesn't want error but got %s", err)
				}

				if test.wantNetwork != pool.network {
					t.Errorf("want global4Counter %d but got %d", test.wantNetwork, pool.network)
				}

				if test.wantMaskHostBits != pool.maskHostBits {
					t.Errorf("want global4MaskBits %d but got %d", test.wantMaskHostBits, pool.maskHostBits)
				}
			},
		)
	}
}

func TestNextIP(t *testing.T) {
	testsSet := []struct {
		description string
		network     string
		wantIP      net.IP
		runBefore   func(*Pool4)
	}{
		{
			description: "first_network_ip",
			network:     "10.0.0.0/24",
			wantIP:      net.IPv4(10, 0, 0, 0),
		},
		{
			description: "last_network_ip",
			network:     "10.0.0.0/24",
			wantIP:      net.IPv4(10, 0, 0, 0),
			runBefore: func(p *Pool4) {
				atomic.StoreUint32(&p.counter, 255)
				p.NextIP()
			},
		},
		{
			description: "no_host_bits",
			network:     "10.0.0.1/32",
			wantIP:      net.IPv4(10, 0, 0, 1),
		},
		{
			description: "no_host_bits_with_wrap_around",
			network:     "10.0.0.1/32",
			wantIP:      net.IPv4(10, 0, 0, 1),
			runBefore: func(p *Pool4) {
				p.NextIP()
			},
		},
	}

	for _, test := range testsSet {
		t.Run(test.description,
			func(t *testing.T) {
				pool, err := NewPool4(test.network)
				if err != nil {
					t.Errorf("doesn't want error but got %s", err)
				}

				if test.runBefore != nil {
					test.runBefore(pool)
				}

				ip := net.IPv4(pool.NextIP())
				if !test.wantIP.Equal(ip) {
					t.Errorf("want ip %s got %s", test.wantIP, ip)
				}
			},
		)
	}
}
