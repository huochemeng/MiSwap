package chain

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"MiSwap/base/evm/eip"
)

const (
	Eth      = "eth"
	Optimism = "optimism"
	Sepolia  = "sepolia"
)

const (
	EthChainID      = 1
	OptimismChainID = 10
	SepoliaChainID  = 11155111
)

func UniformAddress(chainName string, address string) (string, error) {
	switch chainName {
	case "ethereum", "bsc", "polygon":
		addr, err := eip.ToCheckSumAddress(address)
		if err != nil {
			return "", errors.Wrap(err, "failed on uniform evm chain address")
		}
		return strings.ToLower(addr), nil
	case "solana":
		// Solana 地址验证与标准化逻辑
		return validateSolanaAddress(address)
	case "tron":
		// Tron Base58Check 地址验证
		return validateTronAddress(address)
	default:
		return "", fmt.Errorf("unsupported chain: %s", chainName)
	}

}

// todo:后续可添加
func validateTronAddress(address string) (string, error) {
	return "", nil
}

// todo:后续可添加
func validateSolanaAddress(address string) (string, error) {
	return "", nil
}
