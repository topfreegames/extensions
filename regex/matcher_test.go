/*
 * Copyright (c) 2017 TFG Co <backend@tfgco.com>
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

package regex_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/extensions/v9/regex"
)

var _ = Describe("Regex Extension", func() {
	Describe("[Unit]", func() {
		It("should return true for private ips", func() {
			Expect(regex.IsPrivateIP("10.0.1.1")).To(Equal(true))
			Expect(regex.IsPrivateIP("192.168.1.1")).To(Equal(true))
			Expect(regex.IsPrivateIP("172.20.10.13")).To(Equal(true))
		})
		It("should return false for public ips", func() {
			Expect(regex.IsPrivateIP("33.44.55.11")).To(Equal(false))
			Expect(regex.IsPrivateIP("54.44.55.111")).To(Equal(false))
			Expect(regex.IsPrivateIP("182.133.13.44")).To(Equal(false))
		})
	})
})
