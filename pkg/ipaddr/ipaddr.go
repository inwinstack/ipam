/*
Copyright © 2018 inwinSTACK Inc

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ipaddr

import (
	"fmt"
	"net"
	"strings"

	"github.com/thoas/go-funk"

	"github.com/mikioh/ipaddr"
)

type Parser struct {
	Addresses    []string
	AvoidBuggy   bool
	AvoidGateway bool
}

func NewParser(addrs []string, buggy, gateway bool) *Parser {
	return &Parser{
		Addresses:    addrs,
		AvoidBuggy:   buggy,
		AvoidGateway: gateway,
	}
}

func (p *Parser) getIPNets(cidr string) ([]*net.IPNet, error) {
	if !strings.Contains(cidr, "-") {
		_, n, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, fmt.Errorf("invalid CIDR %q", cidr)
		}
		return []*net.IPNet{n}, nil
	}

	fs := strings.SplitN(cidr, "-", 2)
	if len(fs) != 2 {
		return nil, fmt.Errorf("invalid IP range %q", cidr)
	}

	start := net.ParseIP(fs[0])
	if start == nil {
		return nil, fmt.Errorf("invalid IP range %q: invalid start IP %q", cidr, fs[0])
	}

	end := net.ParseIP(fs[1])
	if end == nil {
		return nil, fmt.Errorf("invalid IP range %q: invalid end IP %q", cidr, fs[1])
	}

	var ret []*net.IPNet
	for _, pfx := range ipaddr.Summarize(start, end) {
		n := &net.IPNet{
			IP:   pfx.IP,
			Mask: pfx.Mask,
		}
		ret = append(ret, n)
	}
	return ret, nil
}

func (p *Parser) isGatewayIP(v string) bool {
	gatewayIPs := []string{"1", "254"}
	fs := strings.SplitN(v, ".", 4)
	for _, ip := range gatewayIPs {
		if fs[3] == ip {
			return true
		}
	}
	return false
}

func (p *Parser) isBuggyIP(v string) bool {
	buggyIPs := []string{"0", "255"}
	fs := strings.SplitN(v, ".", 4)
	for _, ip := range buggyIPs {
		if fs[3] == ip {
			return true
		}
	}
	return false
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func (p *Parser) getIPs(ipnet *net.IPNet) []string {
	var ips []string
	for ip := ipnet.IP.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		if p.AvoidBuggy && p.isBuggyIP(ip.String()) {
			continue
		}

		if p.AvoidGateway && p.isGatewayIP(ip.String()) {
			continue
		}
		ips = append(ips, ip.String())
	}
	return ips
}

func (p *Parser) IPs() ([]string, error) {
	var ips []string
	for _, address := range p.Addresses {
		nets, err := p.getIPNets(address)
		if err != nil {
			return nil, fmt.Errorf("Invalid parse CIDR from %+v address", address)
		}

		for _, net := range nets {
			ips = append([]string{}, append(ips, p.getIPs(net)...)...)
		}
	}
	return ips, nil
}

func (p *Parser) FilterIPs(filters ...[]string) ([]string, error) {
	ips, err := p.IPs()
	if err != nil {
		return nil, err
	}

	for _, f := range filters {
		if len(f) > 0 {
			for _, addr := range f {
				ips = funk.FilterString(ips, func(v string) bool { return v != addr })
			}
		}
	}
	return ips, nil
}
