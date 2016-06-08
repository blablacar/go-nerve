package nerve

import (
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"net"
)

func IpLookupNoError(host string, preferIPv4 bool) net.IP {
	ip, err := IpLookup(host, preferIPv4)
	if err != nil {
		logs.WithE(err).WithField("host", host).Error("Host lookup failed, assume localhost can replace it")
		ip = net.IPv4(127, 0, 0, 1)
	}
	return ip
}

func IpLookup(host string, preferIPv4 bool) (net.IP, error) {
	ips, err := net.LookupIP(host)
	if err != nil || len(ips) == 0 || len(ips[0]) == 0 || len(ips[0]) != net.IPv6len {
		return []byte{}, errs.WithEF(err, data.WithField("host", host), "Lookup failed or empty lookup result")
	}

	return processIPs(preferIPv4, ips)
}

func processIPs(preferIpv4 bool, ips []net.IP) (net.IP, error) {
	res := ips[0]
	for _, addr := range ips {
		if preferIpv4 && addr.To4() != nil {
			res = addr
			break
		}
	}
	return res, nil
}

func max(val1 int, val2 int) int {
	if val1 > val2 {
		return val1
	}
	return val2
}
