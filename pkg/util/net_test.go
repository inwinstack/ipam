/*
Copyright © 2018 inwinSTACK.inc

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

package util

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCIDR(t *testing.T) {
	tests := map[string][]string{
		"172.22.132.0/24": {
			"172.22.132.0/24",
		},
		"172.22.132.50-172.22.132.60": {
			"172.22.132.50/31",
			"172.22.132.52/30",
			"172.22.132.56/30",
			"172.22.132.60/32",
		},
	}

	for param, results := range tests {
		nets, err := ParseCIDR(param)
		if err != nil {
			assert.Error(t, err)
		}

		if len(results) != len(nets) {
			assert.Error(t, fmt.Errorf("Wrong parsed nets. Expected %d, got %d", len(results), len(nets)))
		}

		for index, net := range nets {
			assert.Equal(t, results[index], net.String())
		}
	}
}

func TestGetAllIP(t *testing.T) {
	tests := map[string][]string{
		"172.22.132.50/31": {
			"172.22.132.50",
			"172.22.132.51",
		},
		"172.22.132.56/30": {
			"172.22.132.56",
			"172.22.132.57",
			"172.22.132.58",
			"172.22.132.59",
		},
	}

	for param, results := range tests {
		_, net, err := net.ParseCIDR(param)
		if err != nil {
			assert.Error(t, err)
		}

		ips := GetAllIP(net)
		assert.Equal(t, results, ips)
	}
}

func TestParseIPs(t *testing.T) {
	tests := map[string][]string{
		"172.22.132.10,172.22.132.11": {
			"172.22.132.10",
			"172.22.132.11",
		},
	}

	for param, results := range tests {
		ips := ParseIPs(param)
		assert.Equal(t, results, ips)
	}
}
