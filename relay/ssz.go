package relay

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/datachainlab/ethereum-ibc-relay-prover/beacon"
	fastssz "github.com/prysmaticlabs/fastssz"
	"github.com/prysmaticlabs/prysm/v4/encoding/ssz"
)

func generateMerkleProof(leaves [][]byte, generalizedIndex int) ([][]byte, error) {
	if len(leaves) == 0 {
		return nil, fmt.Errorf("leaves length must be greater than 0")
	}
	// check if leaves is a power of 2
	if len(leaves)&(len(leaves)-1) != 0 {
		return nil, fmt.Errorf("leaves length must be a power of 2: actual=%v", len(leaves))
	}
	node, err := fastssz.TreeFromChunks(leaves)
	if err != nil {
		return nil, err
	}
	// NOTE: seems that fastssz sets root as index=1, so the index is off by one from the ethereum consensus-spes
	// https://github.com/ethereum/consensus-specs/blob/c46c3945fd7fbd1226ece1f8d684c4b724b7bdab/ssz/merkle-proofs.md#generalized-merkle-tree-index
	proof, err := node.Prove(generalizedIndex + 1)
	if err != nil {
		return nil, err
	}
	return proof.Hashes, nil
}

func generateExecutionPayloadHeaderProof(header *beacon.ExecutionPayloadHeader, generalizedIndex int) ([][]byte, error) {
	var zero [32]byte
	leaves := [32][]byte{
		header.ParentHash,
		sszBytes(header.FeeRecipient),
		header.StateRoot,
		header.ReceiptsRoot,
		sszBytes(header.LogsBloom),
		header.PrevRandao,
		sszUint64(header.BlockNumber),
		sszUint64(header.GasLimit),
		sszUint64(header.GasUsed),
		sszUint64(header.Timestamp),
		extraDataRootBytes(header.ExtraData),
		header.BaseFeePerGas,
		header.BlockHash,
		header.TransactionsRoot,
		header.WithdrawalsRoot,
		sszUint64(header.BlobGasUsed),
		sszUint64(header.ExcessBlobGas),
		zero[:],
		zero[:],
		zero[:],
		zero[:],
		zero[:],
		zero[:],
		zero[:],
		zero[:],
		zero[:],
		zero[:],
		zero[:],
		zero[:],
		zero[:],
		zero[:],
		zero[:],
	}
	return generateMerkleProof(leaves[:], generalizedIndex)
}

func sszBytes(bz []byte) []byte {
	hh := fastssz.NewHasher()
	hh.PutBytes(bz)
	root, err := hh.HashRoot()
	if err != nil {
		panic(err)
	}
	return root[:]
}

func sszUint64(v uint64) []byte {
	hh := fastssz.NewHasher()
	hh.PutUint64(v)
	root, err := hh.HashRoot()
	if err != nil {
		panic(err)
	}
	return root[:]
}

func extraDataRootBytes(tx []byte) []byte {
	bz, err := extraDataRoot(tx)
	if err != nil {
		panic(err)
	}
	return bz[:]
}

func extraDataRoot(bz []byte) ([32]byte, error) {
	chunkedRoots, err := ssz.PackByChunk([][]byte{bz})
	if err != nil {
		return [32]byte{}, err
	}
	const MAX_EXTRA_DATA_BYTES = 32

	maxLength := (MAX_EXTRA_DATA_BYTES + 31) / 32
	bytesRoot, err := ssz.BitwiseMerkleize(chunkedRoots, uint64(len(chunkedRoots)), uint64(maxLength))
	if err != nil {
		return [32]byte{}, err
	}
	bytesRootBuf := new(bytes.Buffer)
	if err := binary.Write(bytesRootBuf, binary.LittleEndian, uint64(len(bz))); err != nil {
		return [32]byte{}, err
	}
	bytesRootBufRoot := make([]byte, 32)
	copy(bytesRootBufRoot, bytesRootBuf.Bytes())
	return ssz.MixInLength(bytesRoot, bytesRootBufRoot), nil
}
