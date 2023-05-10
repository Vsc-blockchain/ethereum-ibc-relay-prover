package relay

import (
	"math/big"

	lctypes "github.com/datachainlab/ethereum-ibc-relay-prover/light-clients/ethereum/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func (pr *Prover) buildAccountUpdate(blockNumber uint64) (*lctypes.AccountUpdate, error) {
	proof, err := pr.executionClient.GetProof(
		pr.chain.Config().IBCAddress(),
		nil,
		big.NewInt(int64(blockNumber)),
	)
	if err != nil {
		return nil, err
	}
	return &lctypes.AccountUpdate{
		AccountProof:       proof.AccountProofRLP,
		AccountStorageRoot: proof.StorageHash[:],
	}, nil
}

func (pr *Prover) buildStateProof(path []byte, height int64) ([]byte, error) {
	// calculate slot for commitment
	slot := crypto.Keccak256Hash(append(
		crypto.Keccak256Hash(path).Bytes(),
		common.Hash{}.Bytes()...,
	))
	marshaledSlot, err := slot.MarshalText()
	if err != nil {
		return nil, err
	}

	// call eth_getProof
	stateProof, err := pr.executionClient.GetProof(
		pr.chain.Config().IBCAddress(),
		[][]byte{marshaledSlot},
		big.NewInt(height),
	)
	if err != nil {
		return nil, err
	}
	return stateProof.StorageProofRLP[0], nil
}
