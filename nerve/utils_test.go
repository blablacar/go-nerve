package nerve

import (
	. "github.com/onsi/gomega"
	"net"
	"testing"
)

func TestLookup(t *testing.T) {
	RegisterTestingT(t)

	ip4 := net.IPv4(127, 0, 0, 1)
	ip42 := net.IPv4(127, 0, 0, 2)
	ip6 := net.IPv6loopback

	Expect(processIPs(true, []net.IP{ip4})).Should(Equal(ip4))
	Expect(processIPs(true, []net.IP{ip4, ip6})).Should(Equal(ip4))
	Expect(processIPs(true, []net.IP{ip6, ip4})).Should(Equal(ip4))
	Expect(processIPs(true, []net.IP{ip6})).Should(Equal(ip6))

	Expect(processIPs(false, []net.IP{ip4})).Should(Equal(ip4))
	Expect(processIPs(false, []net.IP{ip4, ip6})).Should(Equal(ip4))
	Expect(processIPs(false, []net.IP{ip6, ip4})).Should(Equal(ip6))
	Expect(processIPs(false, []net.IP{ip6})).Should(Equal(ip6))

	Expect(processIPs(false, []net.IP{ip4, ip42})).Should(Equal(ip4))
}
