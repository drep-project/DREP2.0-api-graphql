/*
Package repository implements repository for handling fast and efficient access to data required
by the resolvers of the API server.

Internally it utilizes RPC to access Opera/Lachesis full node for blockchain interaction. Mongo database
for fast, robust and scalable off-chain data storage, especially for aggregated and pre-calculated data mining
results. BigCache for in-memory object storage to speed up loading of frequently accessed entities.
*/package repository

import (
	"fantom-api-graphql/internal/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// StoreNftOwnership stores the given NFT ownership record in persistent storage.
func (p *proxy) StoreNftOwnership(no *types.NftOwnership) error {
	return p.db.StoreNftOwnership(no)
}

// ListNftOwnerships resolves list of nft ownerships based on input data.
func (p *proxy) ListNftOwnerships(contract *common.Address, tokenId *hexutil.Big, owner *common.Address, cursor *string, count int32) (out *types.NftOwnershipList, err error) {
	return p.db.ListNftOwnerships(contract, tokenId, owner, cursor, count)
}