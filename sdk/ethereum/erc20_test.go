package ethereum

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"

	// "math/big"
	"testing"
)

type erc20TestSuite struct {
	suite.Suite
	erc20 IErc20
}

const rpcURL = "http://127.0.0.1:8545"

func TestErc20(t *testing.T) {
	os.Setenv("NSK_BLOCKCHAIN_RPC_URL", rpcURL)
	erc20 := NewErc20Service(nil)
	fmt.Println(erc20.Name("0x224E34A640FC4108FABDb201eD85D909059105fA"))
}

func (s *erc20TestSuite) SetupSuite() {
	os.Setenv("NSK_BLOCKCHAIN_RPC_URL", rpcURL)
	s.erc20 = NewErc20Service(nil)
}

func (s *erc20TestSuite) TearDownSuite() {
}

func (s *erc20TestSuite) TearDownTest() {
}

func (s *erc20TestSuite) methodEqual(body []byte, expected string) {
	value := gjson.GetBytes(body, "method").String()
	s.Require().Equal(expected, value)
}

func (s *erc20TestSuite) paramsEqual(body []byte, expected string) {
	value := gjson.GetBytes(body, "params").Raw
	if expected == "null" {
		s.Require().Equal(expected, value)
	} else {
		s.JSONEq(expected, value)
	}
}

func (s *erc20TestSuite) getBody(request *http.Request) []byte {
	defer request.Body.Close()
	body, err := ioutil.ReadAll(request.Body)
	s.Require().Nil(err)

	return body
}

func (s *erc20TestSuite) TestBalanceOf() {
	_, balance := s.erc20.BalanceOf("0x224E34A640FC4108FABDb201eD85D909059105fA", "0x31ebd457b999bf99759602f5ece5aa5033cb56b3")
	s.Require().Equal("100000000000000000000000", balance.String())
}

func (s *erc20TestSuite) TestAllowanceOf() {
	_, allowance := s.erc20.AllowanceOf("0x224E34A640FC4108FABDb201eD85D909059105fA", "0x1D52a52f5996FDff37317a34EBFbeC7345Be3b55", "0x31ebd457b999bf99759602f5ece5aa5033cb56b3")
	s.Require().Equal("108555083659983933209597798445644913612440610624038028786991485007418559037440", allowance.String())
}

func (s *erc20TestSuite) TestTotalSupply() {
	_, totalSupply := s.erc20.TotalSupply("0x224E34A640FC4108FABDb201eD85D909059105fA")
	s.Require().Equal("1560000000000000000000000000", totalSupply.String())
}

func (s *erc20TestSuite) TestGetSymbol() {
	_, symbol := s.erc20.Symbol("0x224E34A640FC4108FABDb201eD85D909059105fA")
	s.Require().Equal("Hot", symbol)
}

func (s *erc20TestSuite) TestGetName() {
	_, name := s.erc20.Name("0x224E34A640FC4108FABDb201eD85D909059105fA")
	s.Require().Equal("NovaToken", name)
}

func (s *erc20TestSuite) TestGetDecimals() {
	_, decimals := s.erc20.Decimals("0x224E34A640FC4108FABDb201eD85D909059105fA")
	s.Require().Equal(18, decimals)
}

func TestNewErc20Service(t *testing.T) {
	suite.Run(t, new(erc20TestSuite))
}
