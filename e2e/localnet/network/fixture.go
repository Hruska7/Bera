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

package localnet

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"os"
	"path/filepath"
	"strings"

	ginkgo "github.com/onsi/ginkgo/v2"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	cosmlib "pkg.berachain.dev/polaris/cosmos/lib"
	"pkg.berachain.dev/polaris/cosmos/types"
	"pkg.berachain.dev/polaris/eth/common"
	"pkg.berachain.dev/polaris/eth/crypto"
	"pkg.berachain.dev/polaris/lib/encoding"
)

const (
	relativeKeysPath = "../ethkeys/"
	genFilePath      = "../genesis.json"
)

// FixtureConfig is a type defining the configuration of a TestFixture.
type FixtureConfig struct {
	configPath string

	baseImage     string
	containerName string
	httpAddress   string
	wsAdddress    string
	goVersion     string
}

func NewFixtureConfig(configPath, baseImage, containerName, httpAddress, wsAdddress, goVersion string) *FixtureConfig {
	return &FixtureConfig{
		configPath:    configPath,
		baseImage:     baseImage,
		containerName: containerName,
		httpAddress:   httpAddress,
		wsAdddress:    wsAdddress,
		goVersion:     goVersion,
	}
}

// TestFixture is a testing fixture that runs a single Polaris validator node in a Docker container.
type TestFixture struct {
	ContainerizedNode
	t       ginkgo.FullGinkgoTInterface
	keysMap map[string]*ecdsa.PrivateKey
	valAddr common.Address
}

// NewTestFixture creates a new TestFixture.
func NewTestFixture(t ginkgo.FullGinkgoTInterface, config *FixtureConfig) *TestFixture {
	tf := &TestFixture{
		t:       t,
		keysMap: make(map[string]*ecdsa.PrivateKey),
	}

	err := tf.setupTestAccounts(config)
	if err != nil {
		t.Fatal(err)
	}

	tf.ContainerizedNode, err = NewContainerizedNode(
		"localnet",
		"latest",
		config.containerName,
		config.httpAddress,
		config.wsAdddress,
		[]string{
			"GO_VERSION=" + config.goVersion,
			"BASE_IMAGE=" + config.baseImage,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	return tf
}

func (tf *TestFixture) Teardown() error {
	if err := tf.Stop(); err != nil {
		return err
	}
	return tf.Remove()
}

// GenerateTransactOpts generates a new transaction options object for a key by it's name.
func (tf *TestFixture) GenerateTransactOpts(name string) *bind.TransactOpts {
	// Get the nonce from the RPC.
	nonce, err := tf.EthClient().PendingNonceAt(context.Background(), tf.Address(name))
	if err != nil {
		tf.t.Fatal(err)
	}

	// Get the ChainID from the RPC.
	chainID, err := tf.EthClient().ChainID(context.Background())
	if err != nil {
		tf.t.Fatal(err)
	}

	// Build transaction opts object.
	auth, err := bind.NewKeyedTransactorWithChainID(tf.PrivKey(name), chainID)
	if err != nil {
		tf.t.Fatal(err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0) // in wei
	return auth
}

func (tf *TestFixture) PrivKey(name string) *ecdsa.PrivateKey {
	return tf.keysMap[name]
}

func (tf *TestFixture) Address(name string) common.Address {
	privKey, found := tf.keysMap[name]
	if !found {
		return common.Address{}
	}
	return crypto.PubkeyToAddress(privKey.PublicKey)
}

func (tf *TestFixture) ValAddr() common.Address {
	return tf.valAddr
}

// setupTestAccounts loads the test account private keys and validator public key.
func (tf *TestFixture) setupTestAccounts(config *FixtureConfig) error {
	absConfigPath, err := filepath.Abs(config.configPath)
	if err != nil {
		return err
	}

	// read the test account private keys from the keys directory
	keysPath := filepath.Join(absConfigPath, relativeKeysPath)
	keyFiles, err := os.ReadDir(filepath.Clean(keysPath))
	if err != nil {
		return err
	}
	for _, keyFile := range keyFiles {
		keyFileName := keyFile.Name()

		var privKey *ecdsa.PrivateKey
		privKey, err = crypto.LoadECDSA(filepath.Join(keysPath, keyFile.Name()))
		if err != nil {
			return err
		}

		tf.keysMap[strings.Split(keyFileName, ".")[0]] = privKey
	}

	// read the validator public key from the genesis file
	genFile := filepath.Join(absConfigPath, genFilePath)
	genBz, err := os.ReadFile(filepath.Clean(genFile))
	if err != nil {
		return err
	}

	valAddr := encoding.MustUnmarshalJSON[struct {
		AppState struct {
			GenUtil struct {
				GenTxs []struct {
					Body struct {
						Messages []struct {
							ValidatorAddress string `json:"validator_address"`
						} `json:"messages"`
					} `json:"body"`
				} `json:"gen_txs"`
			} `json:"genutil"`
		} `json:"app_state"`
	}](genBz).AppState.GenUtil.GenTxs[0].Body.Messages[0].ValidatorAddress
	types.SetupCosmosConfig()
	acc, err := sdk.ValAddressFromBech32(valAddr)
	if err != nil {
		return err
	}
	tf.valAddr = cosmlib.ValAddressToEthAddress(acc)

	return nil
}
