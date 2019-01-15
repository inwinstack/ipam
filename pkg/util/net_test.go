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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNetworkParser(t *testing.T) {
	tests := []struct {
		NetworkParser *NetworkParser
		IPs           []string
	}{
		{
			NetworkParser: NewNetworkParser([]string{"172.22.132.0/30"}, true, true),
			IPs:           []string{"172.22.132.2", "172.22.132.3"},
		},
		{
			NetworkParser: NewNetworkParser([]string{"172.22.132.0/30"}, true, false),
			IPs:           []string{"172.22.132.1", "172.22.132.2", "172.22.132.3"},
		},
		{
			NetworkParser: NewNetworkParser([]string{"172.22.132.0/30"}, false, true),
			IPs:           []string{"172.22.132.0", "172.22.132.2", "172.22.132.3"},
		},
		{
			NetworkParser: NewNetworkParser([]string{"172.22.132.0/30"}, false, false),
			IPs:           []string{"172.22.132.0", "172.22.132.1", "172.22.132.2", "172.22.132.3"},
		},
	}

	for _, test := range tests {
		ips, err := test.NetworkParser.IPs()
		assert.Nil(t, err)
		assert.Equal(t, test.IPs, ips)
	}
}
