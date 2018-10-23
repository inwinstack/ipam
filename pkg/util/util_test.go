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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func errorGenerator(n int, retryable bool) func() error {
	errorCount := 0
	return func() (err error) {
		if errorCount < n {
			errorCount++
			e := errors.New("testing error")
			if retryable {
				return &RetriableError{Err: e}
			}
			return e
		}
		return nil
	}
}

func TestRetry(t *testing.T) {
	f := errorGenerator(4, true)
	if err := Retry(f, 1, 5); err != nil {
		assert.Fail(t, "Error should not have been raised by retry.")
	}

	f = errorGenerator(5, true)
	if err := Retry(f, 1, 4); err == nil {
		assert.Fail(t, "Error should have been raised by retry.")
	}
}
