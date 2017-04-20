/*
 * Copyright (c) 2016 TFG Co <backend@tfgco.com>
 * Author: TFG Co <backend@tfgco.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of
 * this software and associated documentation files (the "Software"), to deal in
 * the Software without restriction, including without limitation the rights to
 * use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
 * the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package mocks

import (
	"fmt"

	"github.com/go-redis/redis"
)

// RedisMock should be used for tests that need to connect to redis
type RedisMock struct {
	Closed      bool
	Error       error
	Hashs       map[string]map[string]string
	PingCount   int
	PingReponse string
}

// NewRedisMock creates a new redis mock instance
func NewRedisMock(pingResponse string, errOrNil ...error) *RedisMock {
	var err error
	if len(errOrNil) == 1 {
		err = errOrNil[0]
	}
	return &RedisMock{
		Closed:      false,
		Error:       err,
		PingReponse: pingResponse,
	}
}

// Close records that it is closed
func (m *RedisMock) Close() error {
	m.Closed = true

	if m.Error != nil {
		return m.Error
	}

	return nil
}

// HGetAll mocks client.HGetAll
func (m *RedisMock) HGetAll(key string) *redis.StringStringMapCmd {
	if m.Error != nil {
		return redis.NewStringStringMapResult(map[string]string{}, m.Error)
	}
	if val, ok := m.Hashs[key]; ok {
		return redis.NewStringStringMapResult(val, nil)
	}
	return redis.NewStringStringMapResult(map[string]string{}, redis.Nil)
}

// HMSet mocks client.HMSet
func (m *RedisMock) HMSet(key string, fields map[string]interface{}) *redis.StatusCmd {
	if m.Error != nil {
		return redis.NewStatusResult("", m.Error)
	}
	hash := map[string]string{}
	for k, v := range fields {
		if s, ok := v.(string); ok {
			hash[k] = s
		} else {
			hash[k] = fmt.Sprintf("%v", v)
		}
	}
	if m.Hashs != nil {
		m.Hashs[key] = hash
	} else {
		m.Hashs = map[string]map[string]string{
			key: hash,
		}
	}

	return redis.NewStatusResult("OK", nil)
}

// Ping mocks client.Ping
func (m *RedisMock) Ping() *redis.StatusCmd {
	m.PingCount++
	return redis.NewStatusResult(m.PingReponse, m.Error)
}
