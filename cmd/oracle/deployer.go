package oracle

import (
	"context"
)

// ContractDeployer defines the interface for deploying contracts on different chains
type ContractDeployer interface {
	// DownloadContract downloads the contract code from a remote location
	DownloadContract(url string) error

	// DeployContract deploys the contract on chain and returns its address
	DeployContract(
		ctx context.Context,
	) (string, error)

	// Config returns the OracleConfig
	Config() *OracleConfig

	// PrivateKey returns the private key used to deploy the contract
	PrivateKey() string

	// IsContractDeployed returns whether the contract has been deployed to the chain
	IsContractDeployed() (string, bool) // address, bool

	// ContractPath returns the path to the contract file on the local machine
	ContractPath() string

	// ClientConfigPath returns the filepath to the client config file
	ClientConfigPath() string
}
