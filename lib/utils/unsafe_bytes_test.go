// SPDX-License-Identifier: BUSL-1.1
//
// Copyright (C) 2023, Berachain Foundation. All rights reserved.
// Use of this software is govered by the Business Source License included
// in the LICENSE file of this repository and at www.mariadb.com/bsl11.
//
// ANY USE OF THE LICENSED WORK IN VIOLATION OF THIS LICENSE WILL AUTOMATICALLY
// TERMINATE YOUR RIGHTS UNDER THIS LICENSE FOR THE CURRENT AND ALL OTHER
// VERSIONS OF THE LICENSED WORK.
//
// THIS LICENSE DOES NOT GRANT YOU ANY RIGHT IN ANY TRADEMARK OR LOGO OF
// LICENSOR OR ITS AFFILIATES (PROVIDED THAT YOU MAY USE A TRADEMARK OR LOGO OF
// LICENSOR AS EXPRESSLY REQUIRED BY THIS LICENSE).
//
// TO THE EXTENT PERMITTED BY APPLICABLE LAW, THE LICENSED WORK IS PROVIDED ON
// AN “AS IS” BASIS. LICENSOR HEREBY DISCLAIMS ALL WARRANTIES AND CONDITIONS,
// EXPRESS OR IMPLIED, INCLUDING (WITHOUT LIMITATION) WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, NON-INFRINGEMENT, AND
// TITLE.

package utils_test

import (
	"pkg.berachain.dev/stargazer/lib/utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UnsafeStrToBytes", func() {
	When("given a valid string", func() {
		It("should return a byte array with the same content", func() {
			input := "valid string"
			expectedOutput := []byte("valid string")

			output := utils.UnsafeStrToBytes(input)
			Expect(output).To(Equal(expectedOutput))
		})
	})
})

var _ = Describe("UnsafeBytesToStr", func() {
	When("given a valid byte array", func() {
		It("should return a string with the same content", func() {
			input := []byte("valid byte array")
			expectedOutput := "valid byte array"

			output := utils.UnsafeBytesToStr(input)
			Expect(output).To(Equal(expectedOutput))
		})
	})
	When("given empty input", func() {
		It("should return empty string", func() {
			input := []byte{}
			expectedOutput := ""
			output := utils.UnsafeBytesToStr(input)
			Expect(output).To(Equal(expectedOutput))
		})
	})
})
