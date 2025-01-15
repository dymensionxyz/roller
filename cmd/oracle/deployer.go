package oracle

import (
	"context"
	"crypto/ecdsa"
)

// ContractDeployer defines the interface for deploying contracts on different chains
type ContractDeployer interface {
	// DownloadContract downloads the contract code from a remote location
	DownloadContract(url string) error

	// DeployContract deploys the contract on chain and returns its address
	DeployContract(
		ctx context.Context,
		privateKey *ecdsa.PrivateKey,
		contractCode []byte,
	) (string, error)
}
