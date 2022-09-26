// Package resolvers implements GraphQL resolvers to incoming API requests.
package resolvers

import (
	"fantom-api-graphql/internal/repository"
	"fantom-api-graphql/internal/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// NftOwnership represents an NFT ownership
type NftOwnership struct {
	types.NftOwnership
}

// NewNftOwnership creates a new instance of resolvable nft ownership.
func NewNftOwnership(no *types.NftOwnership) *NftOwnership {
	return &NftOwnership{NftOwnership: *no}
}

// Obtained resolves time when NFT was obtained.
func (no NftOwnership) Obtained() hexutil.Uint64 {
	return hexutil.Uint64(no.NftOwnership.Obtained.Unix())
}

// Contract resolves related contract.
func (no NftOwnership) Contract() (*Contract, error) {
	c, err := repository.R().Contract(&no.NftOwnership.Contract)

	if err != nil {
		return nil, err
	}

	return NewContract(c), nil
}
