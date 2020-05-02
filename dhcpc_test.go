package ippool

import (
	"net"
	"testing"
)

func TestIPv4Sequence(t *testing.T) {
	dhcpc, err := New("10.0.0.0/16")
	if err != nil {
		t.Fatal(err)
	}

	wantIPs := make([]net.IP, 256)
	for i := 0; i < len(wantIPs); i++ {
		wantIPs[i] = net.IP{10, 0, 0, byte(i)}
	}

	for i := 0; i < len(wantIPs); i++ {
		ip := dhcpc.IPv4()
		if !wantIPs[i].Equal(ip) {
			t.Fatalf("want ip %q got %q", wantIPs[i], ip)
		}
	}
}

func TestIPv4ExhaustedPool4(t *testing.T) {
	dhcpc, err := New("10.0.0.0/16")
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 65536; i++ {
		ip := dhcpc.IPv4()
		if ip == nil {
			t.Fatalf("want next IP but got nil at iteration %d", i)
		}
	}

	if ip := dhcpc.IPv4(); ip != nil {
		t.Fatalf("expect nil but got %q", ip)
	}
}

func TestIPv4Concurrency(t *testing.T) {
	dhcpc, err := New("10.0.0.0/8")
	if err != nil {
		t.Fatal(err)
	}

	nGoRoutines := 10
	mergeC := make(chan []net.IP)
	for n := 0; n < nGoRoutines; n++ {
		go func() {
			ips := []net.IP{}
			for {
				ip := dhcpc.IPv4()
				if ip == nil {
					mergeC <- ips
					return
				}
				ips = append(ips, ip)
			}
		}()
	}

	mergedIPs := map[string]net.IP{}
	for n := 0; n < nGoRoutines; n++ {
		ips := <-mergeC
		for _, ip := range ips {
			if _, ok := mergedIPs[ip.String()]; ok {
				t.Fatalf("duplicated ip: %s", ip.String())
			}
		}
	}
}

func TestIPv4Release(t *testing.T) {
	dhcpc, err := New("10.0.0.0/16")
	if err != nil {
		t.Fatal(err)
	}

	ip0 := dhcpc.IPv4()
	if ip0.String() != "10.0.0.0" {
		t.Fatalf("unexpected ip %s", ip0)
	}

	dhcpc.genIP4.counter = 0
	ip1 := dhcpc.IPv4()
	if ip1.String() != "10.0.0.1" {
		t.Fatalf("unexpected ip %s", ip1)
	}

	dhcpc.Release4(ip0)
	dhcpc.genIP4.counter = 0
	ip0 = dhcpc.IPv4()
	if ip0.String() != "10.0.0.0" {
		t.Fatalf("unexpected ip %s", ip1)
	}
}
