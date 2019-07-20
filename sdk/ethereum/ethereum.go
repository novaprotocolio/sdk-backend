package ethereum

import (
	"errors"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"

	"crypto/ecdsa"

	"github.com/NovaProtocol/sdk-backend/sdk"
	"github.com/NovaProtocol/sdk-backend/sdk/crypto"
	"github.com/NovaProtocol/sdk-backend/utils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/labstack/gommon/log"
	"github.com/onrik/ethrpc"
	"github.com/shopspring/decimal"
)

var EIP712_DOMAIN_TYPEHASH []byte
var EIP712_ORDER_TYPE []byte

// compile time interface check
var _ sdk.BlockChain = &Ethereum{}
var _ sdk.NovaProtocol = &EthereumNovaProtocol{}
var _ sdk.Nova = &EthereumNova{}

func init() {
	EIP712_DOMAIN_TYPEHASH = crypto.Keccak256([]byte(`EIP712Domain(string name)`))
	EIP712_ORDER_TYPE = crypto.Keccak256([]byte(`Order(address trader,address relayer,address baseToken,address quoteToken,uint256 baseTokenAmount,uint256 quoteTokenAmount,uint256 gasTokenAmount,bytes32 data)`))
}

type EthereumBlock struct {
	*ethrpc.Block
}

func (block *EthereumBlock) Hash() string {
	return block.Block.Hash
}

func (block *EthereumBlock) ParentHash() string {
	return block.Block.ParentHash
}

func (block *EthereumBlock) GetTransactions() []sdk.Transaction {
	txs := make([]sdk.Transaction, 0, 20)

	for i := range block.Block.Transactions {
		tx := block.Block.Transactions[i]
		txs = append(txs, &EthereumTransaction{&tx})
	}

	return txs
}

func (block *EthereumBlock) Number() uint64 {
	return uint64(block.Block.Number)
}

func (block *EthereumBlock) Timestamp() uint64 {
	return uint64(block.Block.Timestamp)
}

type EthereumTransaction struct {
	*ethrpc.Transaction
}

func (t *EthereumTransaction) GetBlockHash() string {
	return t.BlockHash
}

func (t *EthereumTransaction) GetFrom() string {
	return t.From
}

func (t *EthereumTransaction) GetGas() int {
	return t.Gas
}

func (t *EthereumTransaction) GetGasPrice() big.Int {
	return t.GasPrice
}

func (t *EthereumTransaction) GetValue() big.Int {
	return t.Value
}

func (t *EthereumTransaction) GetTo() string {
	return t.To
}

func (t *EthereumTransaction) GetHash() string {
	return t.Hash
}
func (t *EthereumTransaction) GetBlockNumber() uint64 {
	return uint64(*t.BlockNumber)
}

type EthereumTransactionReceipt struct {
	*ethrpc.TransactionReceipt
}

func (r *EthereumTransactionReceipt) GetLogs() (rst []sdk.IReceiptLog) {
	for _, log := range r.Logs {
		l := ReceiptLog{&log}
		rst = append(rst, l)
	}

	return
}

func (r *EthereumTransactionReceipt) GetResult() bool {
	res, err := strconv.ParseInt(r.Status, 0, 64)

	if err != nil {
		panic(err)
	}

	return res == 1
}

func (r *EthereumTransactionReceipt) GetBlockNumber() uint64 {
	return uint64(r.BlockNumber)
}

func (r *EthereumTransactionReceipt) GetBlockHash() string {
	return r.BlockHash
}
func (r *EthereumTransactionReceipt) GetTxHash() string {
	return r.TransactionHash
}
func (r *EthereumTransactionReceipt) GetTxIndex() int {
	return r.TransactionIndex
}

type ReceiptLog struct {
	*ethrpc.Log
}

func (log ReceiptLog) GetRemoved() bool {
	return log.Removed
}

func (log ReceiptLog) GetLogIndex() int {
	return log.LogIndex
}

func (log ReceiptLog) GetTransactionIndex() int {
	return log.TransactionIndex
}

func (log ReceiptLog) GetTransactionHash() string {
	return log.TransactionHash
}

func (log ReceiptLog) GetBlockNum() int {
	return log.BlockNumber
}

func (log ReceiptLog) GetBlockHash() string {
	return log.BlockHash
}

func (log ReceiptLog) GetAddress() string {
	return log.Address
}

func (log ReceiptLog) GetData() string {
	return log.Data
}

func (log ReceiptLog) GetTopics() []string {
	return log.Topics
}

type Ethereum struct {
	client       *ethrpc.EthRPC
	hybridExAddr string
}

func (e *Ethereum) EnableDebug(b bool) {
	e.client.Debug = b
}

func (e *Ethereum) GetBlockByNumber(number uint64) (sdk.Block, error) {

	block, err := e.client.EthGetBlockByNumber(int(number), true)

	if err != nil {
		log.Errorf("get Block by Number failed %+v", err)
		return nil, err
	}

	if block == nil {
		log.Errorf("get Block by Number returns nil block for num: %d", number)
		return nil, errors.New("get Block by Number returns nil block for num: " + strconv.Itoa(int(number)))
	}

	return &EthereumBlock{block}, nil
}

func (e *Ethereum) GetBlockNumber() (uint64, error) {
	number, err := e.client.EthBlockNumber()

	if err != nil {
		log.Errorf("GetBlockNumber failed, %v", err)
		return 0, err
	}

	return uint64(number), nil
}

func (e *Ethereum) GetTransaction(ID string) (sdk.Transaction, error) {
	tx, err := e.client.EthGetTransactionByHash(ID)

	if err != nil {
		log.Errorf("GetTransaction failed, %v", err)
		return nil, err
	}

	return &EthereumTransaction{tx}, nil

}

// this method using ethereum blockchain to send native token
func (e *Ethereum) SendTransaction(to string, amount *big.Int, data []byte, privateKey *ecdsa.PrivateKey) (transactionHash string, err error) {

	from := crypto.PubKey2Address(privateKey.PublicKey)

	recipientAddr := common.HexToAddress(to)
	netVersion, _ := e.client.NetVersion()
	chainID := new(big.Int)
	chainID.SetString(netVersion, 10)
	count, err := e.client.EthGetTransactionCount(from, "pending")
	nonce := uint64(count)
	gasLimit := uint64(100000)
	gasPrice, err := e.client.EthGasPrice()
	if err != nil {
		return
	}

	// this transaction is ethereum transaction
	tx := types.NewTransaction(nonce, recipientAddr, amount, gasLimit, &gasPrice, data)
	signer := types.NewEIP155Signer(chainID)
	signedTx, err := types.SignTx(tx, signer, privateKey)

	if err != nil {
		return
	}

	signedData, _ := rlp.EncodeToBytes(signedTx)

	// fmt.Println(netVersion, data)

	transactionHash, err = e.client.EthSendRawTransaction(common.ToHex(signedData))

	return
}

func (e *Ethereum) GetTransactionReceipt(ID string) (sdk.TransactionReceipt, error) {
	txReceipt, err := e.client.EthGetTransactionReceipt(ID)

	if err != nil {
		log.Errorf("GetTransactionReceipt failed, %v", err)
		return nil, err
	}

	return &EthereumTransactionReceipt{txReceipt}, nil
}

func (e *Ethereum) GetTransactionAndReceipt(ID string) (sdk.Transaction, sdk.TransactionReceipt, error) {
	txReceiptChannel := make(chan sdk.TransactionReceipt)

	go func() {
		rec, _ := e.GetTransactionReceipt(ID)
		txReceiptChannel <- rec
	}()

	txInfoChannel := make(chan sdk.Transaction)
	go func() {
		tx, _ := e.GetTransaction(ID)
		txInfoChannel <- tx
	}()

	return <-txInfoChannel, <-txReceiptChannel, nil
}

func (e *Ethereum) GetTokenBalance(tokenAddress, address string) decimal.Decimal {
	res, err := e.client.EthCall(ethrpc.T{
		To:   tokenAddress,
		From: address,
		Data: fmt.Sprintf("%s000000000000000000000000%s", ERC20BalanceOf, without0xPrefix(address)),
	}, "latest")

	if err != nil {
		panic(err)
	}

	return utils.StringToDecimal(res)
}

func without0xPrefix(address string) string {
	if address[:2] == "0x" {
		address = address[2:]
	}

	return address
}

func (e *Ethereum) GetTokenAllowance(tokenAddress, proxyAddress, address string) decimal.Decimal {
	res, err := e.client.EthCall(ethrpc.T{
		To:   tokenAddress,
		From: address,
		Data: fmt.Sprintf("%s000000000000000000000000%s000000000000000000000000%s", ERC20Allowance, without0xPrefix(address), without0xPrefix(proxyAddress)),
	}, "latest")

	if err != nil {
		panic(err)
	}

	return utils.StringToDecimal(res)
}

func (e *Ethereum) GetHotFeeDiscount(address string) decimal.Decimal {
	if address == "" {
		return decimal.New(1, 0)
	}

	from := address

	res, err := e.client.EthCall(ethrpc.T{
		To:   e.hybridExAddr,
		From: from,
		Data: fmt.Sprintf("%s000000000000000000000000%s", DiscountedRate, without0xPrefix(address)),
	}, "latest")

	if err != nil {
		panic(err)
	}

	return utils.StringToDecimal(res).Div(decimal.New(1, 2))
}

func (e *Ethereum) IsValidSignature(address string, message string, signature string) (bool, error) {
	if len(address) != 42 {
		return false, errors.New("address must be 42 size long")
	}

	if len(signature) != 132 {
		return false, errors.New("signature must be 132 size long")
	}

	var hashBytes []byte
	if strings.HasPrefix(message, "0x") {
		hashBytes = utils.Hex2Bytes(message[2:])
	} else {
		hashBytes = []byte(message)
	}

	signatureByte := utils.Hex2Bytes(signature[2:])
	pk, err := crypto.PersonalEcRecover(hashBytes, signatureByte)

	if err != nil {
		return false, err
	}

	return "0x"+strings.ToLower(pk) == strings.ToLower(address), nil
}

func (e *Ethereum) SendRawTransaction(tx interface{}) (string, error) {
	rawTransaction := tx.(string)
	return e.client.EthSendRawTransaction(rawTransaction)
}

func (e *Ethereum) GetTransactionCount(address string) (int, error) {
	return e.client.EthGetTransactionCount(address, "latest")
}

func NewEthereum(rpcUrl string, hybridExAddr string) *Ethereum {
	if hybridExAddr == "" {
		hybridExAddr = os.Getenv("NSK_HYBRID_EXCHANGE_ADDRESS")
	}

	if hybridExAddr == "" {
		panic(fmt.Errorf("NewEthereum need argument hybridExAddr"))
	}

	return &Ethereum{
		client:       ethrpc.New(rpcUrl),
		hybridExAddr: hybridExAddr,
	}
}

func IsValidSignature(address string, message string, signature string) (bool, error) {
	return new(Ethereum).IsValidSignature(address, message, signature)
}

func PersonalSign(message []byte, privateKey string) ([]byte, error) {
	return crypto.PersonalSign(message, privateKey)
}

// orderbook methods
func (e *Ethereum) GetOrderbook(pairName, orderID string) []byte {	
	result, _ := e.client.Call("orderbook_getOrder", pairName, orderID)
	bytes, _ := result.MarshalJSON()
	return bytes
}

func (e *Ethereum) ProcessOrderbook(payload map[string]string) []byte {		
	result, _ := e.client.Call("orderbook_processOrder", payload)
	bytes, _ := result.MarshalJSON()
	return bytes
}

func (e *Ethereum) CancelOrderbook(payload map[string]string) bool {
	result, _ := e.client.Call("orderbook_cancelOrder", payload)	
	return result == nil
}

func (e *Ethereum) BestAskList(pairName string) []byte {
	result, _ := e.client.Call("orderbook_getBestAskList", pairName)	
	bytes, _ := result.MarshalJSON()
	return bytes
}

func (e *Ethereum) BestBidList(pairName string) []byte {
	result, _ := e.client.Call("orderbook_getBestBidList", pairName)	
	bytes, _ := result.MarshalJSON()
	return bytes
}