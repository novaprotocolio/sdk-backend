package sdk

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
)

type MockBlockchain struct {
	BlockChain
	mock.Mock
}

type MockNovaProtocol struct {
	NovaProtocol
	mock.Mock
}

type MockNova struct {
	NovaProtocol
	BlockChain
}

func NewMockNova() *MockNova {
	return &MockNova{
		&MockNovaProtocol{},
		&MockBlockchain{},
	}
}

func (m *MockBlockchain) GetBlockNumber() (uint64, error) {
	args := m.Called()
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockBlockchain) GetBlockByNumber(blockNumber uint64) (Block, error) {
	args := m.Called(blockNumber)
	return args.Get(0).(Block), args.Error(1)
}

func (m *MockBlockchain) GetTransaction(ID string) (Transaction, error) {
	args := m.Called(ID)
	return args.Get(0).(Transaction), args.Error(1)
}

func (m *MockBlockchain) GetTransactionReceipt(ID string) (TransactionReceipt, error) {
	args := m.Called(ID)
	return args.Get(0).(TransactionReceipt), args.Error(1)
}

func (m *MockBlockchain) GetTransactionAndReceipt(ID string) (Transaction, TransactionReceipt, error) {
	args := m.Called(ID)
	return args.Get(0).(Transaction), args.Get(1).(TransactionReceipt), args.Error(2)
}

func (m *MockBlockchain) GetTokenBalance(tokenAddress string, address string) decimal.Decimal {
	args := m.Called(tokenAddress, address)
	return args.Get(0).(decimal.Decimal)
}

func (m *MockBlockchain) GetTokenAllowance(tokenAddress, proxyAddress, address string) decimal.Decimal {
	args := m.Called(tokenAddress, address)
	return args.Get(0).(decimal.Decimal)
}

func (m *MockBlockchain) GetHotFeeDiscount(address string) decimal.Decimal {
	args := m.Called(address)
	return args.Get(0).(decimal.Decimal)
}

//func (m *MockBlockchainClient) GetOrderHash(order *OrderParam, addressSet OrderAddressSet, novaContractAddress string) []byte {
//	args := m.Called(order, addressSet, novaContractAddress)
//	return args.Get(0).([]byte)
//}

func (m *MockBlockchain) IsValidSignature(address string, message string, signature string) (bool, error) {
	args := m.Called(address, message, signature)
	return args.Bool(0), args.Error(1)
}

func (m *MockBlockchain) SendTransaction(to string, amount *big.Int, data []byte, privateKey *ecdsa.PrivateKey) (transactionHash string, err error) {
	args := m.Called(to, amount, data, privateKey)
	return args.String(0), args.Error(1)
}

func (m *MockBlockchain) SendRawTransaction(tx interface{}) (string, error) {
	args := m.Called(tx)
	return args.String(0), args.Error(1)
}

func (m *MockNovaProtocol) GenerateOrderData(version, expiredAtSeconds, salt int64, asMakerFeeRate, asTakerFeeRate, makerRebateRate decimal.Decimal, isSell, isMarket, isMakerOnly bool) string {
	args := m.Called(version, expiredAtSeconds, salt, asMakerFeeRate, asTakerFeeRate, makerRebateRate, isSell, isMarket, isMakerOnly)
	return args.String(0)
}

func (m *MockNovaProtocol) GetOrderHash(order *Order) []byte {
	args := m.Called(order)
	return args.Get(0).([]byte)
}

func (m *MockNovaProtocol) GetMatchOrderCallData(takerOrder *Order, makerOrders []*Order, baseTokenFilledAmounts []*big.Int) []byte {
	args := m.Called(takerOrder, makerOrders, baseTokenFilledAmounts)
	return args.Get(0).([]byte)
}
