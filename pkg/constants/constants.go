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

package constants

const (
	// DefaultPool represents the default pool name for a namespace.
	DefaultPool = "default"
	// DefaultNumberOfIP represents the default number of IP for a namespace.
	DefaultNumberOfIP = 1
	// AnnKeyIPs will set in namespace resource to display IPs of allocation.
	AnnKeyIPs = "inwinstack.com/allocated-ips"
	// AnnKeyLatestIP will set in namespace resource to display the latest allocated IP.
	AnnKeyLatestIP = "inwinstack.com/allocated-latest-ip"
	// AnnKeyNumberOfIP will set in namespace resource to represent the number of IP want to allocate.
	AnnKeyNumberOfIP = "inwinstack.com/allocate-ip-number"
	// AnnKeyPoolName will set in namespace resource to represent the current IP pool.
	AnnKeyPoolName = "inwinstack.com/allocate-pool-name"
	// AnnKeyNamespaceRefresh will be set in namespace resource to refresh annotations.
	AnnKeyNamespaceRefresh = "inwinstack.com/namespace-refresh"
)
