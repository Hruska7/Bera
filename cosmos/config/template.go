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

//nolint:lll,cyclop // template file.
package config

const (
	PolarisConfigTemplate = `
###############################################################################
###                                 Polaris                                 ###
###############################################################################
# General Polaris settings
[polaris]

[polaris.polar]
# Gas cap for RPC requests
rpc-gas-cap = "{{ .Polaris.Polar.RPCGasCap }}"

# Timeout setting for EVM operations via RPC
rpc-evm-timeout = "{{ .Polaris.Polar.RPCEVMTimeout }}"

# Transaction fee cap for RPC requests
rpc-tx-fee-cap = "{{ .Polaris.Polar.RPCTxFeeCap }}"


# Gas price oracle settings for Polaris
[polaris.polar.gpo]
# Number of blocks to check for gas prices
blocks = {{ .Polaris.Polar.GPO.Blocks }}

# Percentile of gas price to use
percentile = {{ .Polaris.Polar.GPO.Percentile }}

# Maximum header history for gas price determination
max-header-history = {{ .Polaris.Polar.GPO.MaxHeaderHistory }}

# Maximum block history for gas price determination
max-block-history = {{ .Polaris.Polar.GPO.MaxBlockHistory }}

# Default gas price value
default = "{{ .Polaris.Polar.GPO.Default }}"

# Maximum gas price value
max-price = "{{ .Polaris.Polar.GPO.MaxPrice }}"

# Prices to ignore for gas price determination
ignore-price = "{{ .Polaris.Polar.GPO.IgnorePrice }}"


# Node-specific settings
[polaris.node]
# Name of the node
name = "{{ .Polaris.Node.Name }}"

# User identity associated with the node
user-ident = "{{ .Polaris.Node.UserIdent }}"

# Version of the node
version = "{{ .Polaris.Node.Version }}"

# Directory for storing node data
data-dir = "{{ .Polaris.Node.DataDir }}"

# Directory for storing node keys
key-store-dir = "{{ .Polaris.Node.KeyStoreDir }}"

# Path to the external signer
external-signer = "{{ .Polaris.Node.ExternalSigner }}"

# Whether to use lightweight KDF
use-lightweight-kdf = {{ .Polaris.Node.UseLightweightKDF }}

# Allow insecure unlock
insecure-unlock-allowed = {{ .Polaris.Node.InsecureUnlockAllowed }}

# USB setting for the node
usb = {{ .Polaris.Node.USB }}

# Path to smart card daemon
smart-card-daemon-path = "{{ .Polaris.Node.SmartCardDaemonPath }}"

# IPC path for the node
ipc-path = "{{ .Polaris.Node.IPCPath }}"

# Host for HTTP requests
http-host = "{{ .Polaris.Node.HTTPHost }}"

# Port for HTTP requests
http-port = {{ .Polaris.Node.HTTPPort }}

# CORS settings for HTTP
http-cors = [{{ range $index, $element := .Polaris.Node.HTTPCors }}{{ if $index }}, {{ end }}"{{ $element }}"{{ end }}]

# Virtual hosts for HTTP
http-virtual-hosts = [{{ range $index, $element := .Polaris.Node.HTTPVirtualHosts }}{{ if $index }}, {{ end }}"{{ $element }}"{{ end }}]

# Enabled modules for HTTP
http-modules = [{{ range $index, $element := .Polaris.Node.HTTPModules }}{{ if $index }}, {{ end }}"{{ $element }}"{{ end }}]

# Path prefix for HTTP
http-path-prefix = "{{ .Polaris.Node.HTTPPathPrefix }}"

# Address for authentication
auth-addr = "{{ .Polaris.Node.AuthAddr }}"

# Port for authentication
auth-port = {{ .Polaris.Node.AuthPort }}

# Virtual hosts for authentication
auth-virtual-hosts = [{{ range $index, $element := .Polaris.Node.AuthVirtualHosts }}{{ if $index }}, {{ end }}"{{ $element }}"{{ end }}]

# Host for WebSockets
ws-host = "{{ .Polaris.Node.WSHost }}"

# Port for WebSockets
ws-port = {{ .Polaris.Node.WSPort }}

# Path prefix for WebSockets
ws-path-prefix = "{{ .Polaris.Node.WSPathPrefix }}"

# Origins allowed for WebSockets
ws-origins = [{{ range $index, $element := .Polaris.Node.WSOrigins }}{{ if $index }}, {{ end }}"{{ $element }}"{{ end }}]

# Enabled modules for WebSockets
ws-modules = [{{ range $index, $element := .Polaris.Node.WSModules }}{{ if $index }}, {{ end }}"{{ $element }}"{{ end }}]

# Expose all settings for WebSockets
ws-expose-all = {{ .Polaris.Node.WSExposeAll }}

# CORS settings for GraphQL
graphql-cors = [{{ range $index, $element := .Polaris.Node.GraphQLCors }}{{ if $index }}, {{ end }}"{{ $element }}"{{ end }}]

# Virtual hosts for GraphQL
graphql-virtual-hosts = [{{ range $index, $element := .Polaris.Node.GraphQLVirtualHosts }}{{ if $index }}, {{ end }}"{{ $element }}"{{ end }}]

# Allow unprotected transactions
allow-unprotected-txs = {{ .Polaris.Node.AllowUnprotectedTxs }}

# Limit for batch requests
batch-request-limit = {{ .Polaris.Node.BatchRequestLimit }}

# Maximum size for batch responses
batch-response-max-size = {{ .Polaris.Node.BatchResponseMaxSize }}

# JWT secret for authentication
jwt-secret = "{{ .Polaris.Node.JWTSecret }}"

# Database engine for the node
db-engine = "{{ .Polaris.Node.DBEngine }}"


# HTTP timeout settings for the node
[polaris.node.http-timeouts]
# Timeout for reading HTTP requests
read-timeout = "{{ .Polaris.Node.HTTPTimeouts.ReadTimeout }}"

# Timeout for reading HTTP request headers
read-header-timeout = "{{ .Polaris.Node.HTTPTimeouts.ReadHeaderTimeout }}"

# Timeout for writing HTTP responses
write-timeout = "{{ .Polaris.Node.HTTPTimeouts.WriteTimeout }}"

# Timeout for idle HTTP connections
idle-timeout = "{{ .Polaris.Node.HTTPTimeouts.IdleTimeout }}"
`
)
