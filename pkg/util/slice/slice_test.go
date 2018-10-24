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

package slice

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveItem(t *testing.T) {
	expected := []string{"1", "2", "3", "4"}
	tests := []string{"1", "2", "3", "4", "5"}
	assert.Equal(t, expected, RemoveItem(tests, "5"))
}

func TestRemoveItems(t *testing.T) {
	expected := []string{"1", "2"}
	tests := []string{"1", "2", "3", "4", "5"}
	assert.Equal(t, expected, RemoveItems(tests, []string{"3", "4", "5"}))
}

func TestUnique(t *testing.T) {
	expected := []string{"1", "2", "3", "4", "5"}
	tests := []string{"1", "2", "3", "4", "5", "5", "4", "3"}
	assert.Equal(t, expected, Unique(tests))
}

func TestContains(t *testing.T) {
	expected := map[string]bool{
		"1": true,
		"2": false,
		"3": true,
	}

	tests := []string{"1", "3", "4", "5"}
	for param, result := range expected {
		assert.Equal(t, result, Contains(tests, param))
	}
}
