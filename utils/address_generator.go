package utils

import (
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/tyler-smith/go-bip32"
)

type AddressGenerator struct {
	masterPublicKey *bip32.Key
}

var EmptyAddress = common.HexToAddress("0x0")

// NewAddressGenerator : generate new address from master key : cfg.Ethereum.MasterPublicKey
func NewAddressGenerator(masterPublicKeyString string) (*AddressGenerator, error) {
	deserializedMasterPublicKey, err := bip32.B58Deserialize(masterPublicKeyString)
	if err != nil {
		return nil, errors.New("Error deserializing master public key: " + err.Error())
	}

	if deserializedMasterPublicKey.IsPrivate {
		return nil, errors.New("Key is not a master public key")
	}

	return &AddressGenerator{deserializedMasterPublicKey}, nil
}

// common.Address is pointer already, it is slice
func (g *AddressGenerator) Generate(index uint64) (common.Address, error) {
	if g.masterPublicKey == nil {
		return EmptyAddress, errors.New("No master public key set")
	}

	accountKey, err := g.masterPublicKey.NewChildKey(uint32(index))
	if err != nil {
		return EmptyAddress, errors.New("Error creating new child key: " + err.Error())
	}

	x, y := secp256k1.DecompressPubkey(accountKey.Key)

	uncompressed := make([]byte, 64)
	copy(uncompressed[0:32], x.Bytes())
	copy(uncompressed[32:], y.Bytes())

	keccak := crypto.Keccak256(uncompressed)
	address := common.BytesToAddress(keccak[12:]) // Encode lower 160 bits/20 bytes
	return address, nil
}
