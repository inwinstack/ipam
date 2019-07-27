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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIPs(t *testing.T) {
	tests := []struct {
		Parser *Parser
		IPs    []string
	}{
		{
			Parser: NewParser([]string{"172.22.132.0/30"}, true, true),
			IPs:    []string{"172.22.132.2", "172.22.132.3"},
		},
		{
			Parser: NewParser([]string{"172.22.132.0/30"}, true, false),
			IPs:    []string{"172.22.132.1", "172.22.132.2", "172.22.132.3"},
		},
		{
			Parser: NewParser([]string{"172.22.132.0/30"}, false, true),
			IPs:    []string{"172.22.132.0", "172.22.132.2", "172.22.132.3"},
		},
		{
			Parser: NewParser([]string{"172.22.132.0/30"}, false, false),
			IPs:    []string{"172.22.132.0", "172.22.132.1", "172.22.132.2", "172.22.132.3"},
		},
	}

	for _, test := range tests {
		ips, err := test.Parser.IPs()
		assert.Nil(t, err)
		assert.Equal(t, test.IPs, ips)
	}
}

func TestFilterIPs(t *testing.T) {
	tests := []struct {
		Parser  *Parser
		Filters []string
		IPs     []string
	}{
		{
			Parser:  NewParser([]string{"172.22.132.0-172.22.132.5"}, true, true),
			Filters: []string{"172.22.132.4", "172.22.132.5"},
			IPs:     []string{"172.22.132.2", "172.22.132.3"},
		},
		{
			Parser:  NewParser([]string{"172.22.132.0/30"}, true, false),
			Filters: []string{"172.22.132.3"},
			IPs:     []string{"172.22.132.1", "172.22.132.2"},
		},
	}

	for _, test := range tests {
		ips, err := test.Parser.FilterIPs(test.Filters)
		assert.Nil(t, err)
		assert.Equal(t, test.IPs, ips)
	}
}
