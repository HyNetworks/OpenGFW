package geo

import (
	"net"
	"testing"

	"github.com/apernet/OpenGFW/ruleset/builtins/geo/v2geo"
)

func TestGeoIPMatcherSortsNetworks(t *testing.T) {
	list := &v2geo.GeoIP{Cidr: []*v2geo.CIDR{
		{Ip: []byte{192, 0, 2, 0}, Prefix: 24},
		{Ip: []byte{10, 0, 0, 0}, Prefix: 24},
	}}
	matcher, err := newGeoIPMatcher(list)
	if err != nil {
		t.Fatal(err)
	}
	if got := matcher.N4[0].IP.String(); got != "10.0.0.0" {
		t.Fatalf("first network = %s, want 10.0.0.0", got)
	}
	for _, ip := range []net.IP{net.IPv4(10, 0, 0, 42), net.IPv4(192, 0, 2, 42)} {
		if !matcher.Match(HostInfo{IPv4: ip}) {
			t.Errorf("expected %s to match", ip)
		}
	}
	if matcher.Match(HostInfo{IPv4: net.IPv4(203, 0, 113, 42)}) {
		t.Fatal("unexpected match outside configured networks")
	}
}
