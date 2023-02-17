// Copyright (C) 2023, Berachain Foundation. All rights reserved.
// See the file LICENSE for licensing terms.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
// CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
// OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package core_test

import (
	"context"
	"math/big"

	"github.com/berachain/stargazer/eth/common"
	"github.com/berachain/stargazer/eth/core"
	"github.com/berachain/stargazer/eth/core/mock"
	"github.com/berachain/stargazer/eth/core/types"
	"github.com/berachain/stargazer/eth/core/vm"
	vmmock "github.com/berachain/stargazer/eth/core/vm/mock"
	"github.com/berachain/stargazer/eth/crypto"
	"github.com/berachain/stargazer/eth/params"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	william = common.HexToAddress("0x123")
	key, _  = crypto.GenerateEthKey()
	signer  = types.LatestSignerForChainID(params.DefaultChainConfig.ChainID)

	legacyTxData = &types.LegacyTx{
		Nonce:    0,
		To:       &william,
		Gas:      100000,
		GasPrice: big.NewInt(2),
		Data:     []byte("abcdef"),
	}
)

var _ = Describe("StateProcessor", func() {
	var (
		// evm         *vmmock.StargazerEVMMock
		sdb *vmmock.StargazerStateDBMock
		// msg         *mock.MessageMock
		host          *mock.StargazerHostChainMock
		bp            *mock.BlockPluginMock
		gp            *mock.GasPluginMock
		cp            *mock.ConfigurationPluginMock
		pp            *mock.PrecompilePluginMock
		sp            *core.StateProcessor
		blockNumber   uint64
		blockGasLimit uint64
	)

	BeforeEach(func() {
		// evm = vmmock.NewStargazerEVM()
		sdb = vmmock.NewEmptyStateDB()
		// msg = mock.NewEmptyMessage()
		host = mock.NewMockHost()
		bp = &mock.BlockPluginMock{}
		gp = mock.NewGasPluginMock()
		cp = &mock.ConfigurationPluginMock{}
		pp = &mock.PrecompilePluginMock{}
		host.GetBlockPluginFunc = func() core.BlockPlugin {
			return bp
		}
		host.GetGasPluginFunc = func() core.GasPlugin {
			return gp
		}
		host.GetConfigurationPluginFunc = func() core.ConfigurationPlugin {
			return cp
		}
		host.GetPrecompilePluginFunc = func() core.PrecompilePlugin {
			return pp
		}
		sp = core.NewStateProcessor(host, sdb, vm.Config{}, true)
		Expect(sp).ToNot(BeNil())
		blockNumber = params.DefaultChainConfig.LondonBlock.Uint64() + 1
		blockGasLimit = 1000000

		bp.PrepareFunc = func(ctx context.Context) {
			// no-op
		}
		bp.GetStargazerHeaderAtHeightFunc = func(height int64) *types.StargazerHeader {
			header := types.NewEmptyStargazerHeader()
			header.GasLimit = blockGasLimit
			header.BaseFee = big.NewInt(1)
			header.Coinbase = common.BytesToAddress([]byte{2})
			header.Number = big.NewInt(int64(blockNumber))
			header.Time = uint64(3)
			return header
		}
		cp.PrepareFunc = func(ctx context.Context) {
			// no-op
		}
		cp.ChainConfigFunc = func() *params.ChainConfig {
			return params.DefaultChainConfig
		}
		cp.ExtraEipsFunc = func() []int {
			return []int{}
		}
		pp.HasFunc = func(addr common.Address) bool {
			return false
		}

		gp.SetBlockGasLimit(blockGasLimit)
		sp.Prepare(context.Background(), 0)
	})

	Context("Empty block", func() {
		It("should build a an empty block", func() {
			block, err := sp.Finalize(context.Background())
			Expect(err).To(BeNil())
			Expect(block).ToNot(BeNil())
			Expect(block.TxIndex()).To(Equal(uint(0)))
		})
	})

	Context("Block with transactions", func() {
		BeforeEach(func() {
			_, err := sp.Finalize(context.Background())
			Expect(err).To(BeNil())

			pp.ResetFunc = func(ctx context.Context) {
				// no-op
			}

			sp.Prepare(context.Background(), int64(blockNumber))
		})

		It("should error on an unsigned transaction", func() {
			receipt, err := sp.ProcessTransaction(context.Background(), types.NewTx(legacyTxData))
			Expect(err).ToNot(BeNil())
			Expect(receipt).To(BeNil())
			block, err := sp.Finalize(context.Background())
			Expect(err).To(BeNil())
			Expect(block).ToNot(BeNil())
			Expect(block.TxIndex()).To(Equal(uint(0)))
		})

		It("should not error on a signed transaction", func() {
			signedTx := types.MustSignNewTx(key, signer, legacyTxData)
			result, err := sp.ProcessTransaction(context.Background(), signedTx)
			Expect(err).To(BeNil())
			Expect(result).ToNot(BeNil())
			Expect(result.Status).To(Equal(types.ReceiptStatusSuccessful))
			Expect(result.BlockNumber).To(Equal(big.NewInt(int64(blockNumber))))
			Expect(result.TransactionIndex).To(Equal(uint(0)))
			Expect(result.TxHash.Hex()).To(Equal(signedTx.Hash().Hex()))
			Expect(result.GasUsed).ToNot(BeZero())
			block, err := sp.Finalize(context.Background())
			Expect(err).To(BeNil())
			Expect(block).ToNot(BeNil())
			Expect(block.TxIndex()).To(Equal(uint(1)))
		})

		It("should add a contract address to the receipt", func() {
			legacyTxDataCopy := *legacyTxData
			legacyTxDataCopy.To = nil
			signedTx := types.MustSignNewTx(key, signer, &legacyTxDataCopy)
			result, err := sp.ProcessTransaction(context.Background(), signedTx)
			Expect(err).To(BeNil())
			Expect(result).ToNot(BeNil())
			Expect(result.ContractAddress).ToNot(BeNil())
			block, err := sp.Finalize(context.Background())
			Expect(err).To(BeNil())
			Expect(block).ToNot(BeNil())
			Expect(block.TxIndex()).To(Equal(uint(1)))
		})

		It("should mark a receipt with a virtual machine error as failed", func() {
			// sdb.GetCodeFunc = func(addr common.Address) []byte {
			// 	return solidity.RevertableTx.Bin
			// }
			// sdb.GetCodeHashFunc = func(addr common.Address) common.Hash {
			// 	return crypto.Keccak256Hash(solidity.RevertableTx.Bin)
			// }
			// signedTx := types.MustSignNewTx(key, signer, legacyTxData)
			// result, err := sp.ProcessTransaction(context.Background(), signedTx)
			// Expect(err).To(BeNil())
			// Expect(result).ToNot(BeNil())
			// Expect(result.Status).To(Equal(types.ReceiptStatusFailed))
			// block, err := sp.Finalize(context.Background())
			// Expect(err).To(BeNil())
			// Expect(block).ToNot(BeNil())
			// Expect(block.TxIndex()).To(Equal(uint(1)))
		})

		It("should not include consensus breaking transactions", func() {
			// signedTx := types.MustSignNewTx(key, signer, legacyTxData)
			// result, err := sp.ProcessTransaction(context.Background(), signedTx)
			// Expect(err).To(BeNil())
			// Expect(result).ToNot(BeNil())
			// Expect(result.Status).To(Equal(types.ReceiptStatusFailed))
			// block, err := sp.Finalize(context.Background(), blockNumber)
			// Expect(err).To(BeNil())
			// Expect(block).ToNot(BeNil())
			// Expect(len(block.Transactions)).To(Equal(1))
		})
	})
})

var _ = Describe("EVM Test Suite", func() {
	// var host *mock.StargazerHostChainMock

	// hash1 := common.Hash{1}
	// hash2 := common.Hash{2}
	// hash3 := common.Hash{3}
	// hash4 := common.Hash{4}

	// currentHeader := &types.StargazerHeader{
	// 	Header: &types.Header{
	// 		Number:     big.NewInt(int64(123)),
	// 		BaseFee:    big.NewInt(69),
	// 		ParentHash: common.Hash{111},
	// 	},
	// 	HostHash: common.Hash{1},
	// }

	Context("TestGetHashFunc", func() {
		BeforeEach(func() {
			// host = mock.NewMockHost()
		})
		// It("should return the correct hash", func() {
		// 	host.StargazerHeaderAtHeightFunc = func(ctx context.Context, height uint64) *types.StargazerHeader {
		// 		return &types.StargazerHeader{
		// 			Header: &types.Header{
		// 				Number:     big.NewInt(int64(height)),
		// 				BaseFee:    big.NewInt(69),
		// 				ParentHash: common.Hash{123},
		// 			},
		// 			HostHash: crypto.Keccak256Hash([]byte{byte(height)}),
		// 		}
		// 	}
		// 	hash := core.GetHashFn(context.Background(), currentHeader, host)
		// 	Expect(hash(112)).To(Equal(crypto.Keccak256Hash([]byte{byte(112)})))
		// })
	})
})
