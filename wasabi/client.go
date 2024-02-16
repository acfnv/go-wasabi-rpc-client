// Package wasabi provides a client for the wasabi daemon via RPC.
package wasabi

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"net"
	"net/http"
	"sync"
)

// Client is a wasabi-wallet-rpc client.
type Client interface {
	// IsWasabiWalletUp checks if Wasabi is running and reachable.
	IsWasabiWalletUp() bool

	// GetStatus returns information useful to understand Wasabi and its synchronization status.
	GetStatus() (GetStatusResponse, error)

	// CreateWallet creates a new wallet with the given name and password and returns the twelve recovery words of the freshly generated wallet in one string (space separated).
	CreateWallet(walletName string, password string) (string, error)

	// LoadWallet loads a wallet with the given name. Before accessing the wallet for the first time, it must be loaded.
	LoadWallet(walletName string) error

	// ListCoins returns the list of previously spent and currently unspent coins (confirmed and unconfirmed).
	ListCoins(walletName string) ([]ListCoinsResponse, error)

	// ListUnspentCoins returns the list of confirmed and unconfirmed coins that are unspent.
	ListUnspentCoins(walletName string) ([]ListCoinsResponse, error)

	// GetWalletInfo returns information about the current loaded wallet.
	GetWalletInfo(walletName string) (GetWalletInfoResponse, error)

	// GetNewAddress creates an address and returns detailed information about it.
	GetNewAddress(walletName string, label string) (GetNewAddressResponse, error)

	// Send builds and broadcasts a transaction.
	Send(walletName string, payments []Payment, coins []Coin, feeTarget int, password string) (SendResponse, error)

	// Build builds a transaction. It is similar to the send method, except that it will not automatically broadcast the transaction. So it is also possible to send to many and to subtract the fee.
	Build(walletName string, payments []Payment, coins []Coin, feeTarget int, password string) (string, error)

	// Broadcast broadcasts a transaction. Enter the transaction hex in the params field. Returns the transaction id.
	Broadcast(walletName string, hex string) (string, error)

	// GetHistory returns the list of all transactions sent and received.
	GetHistory(walletName string) ([]Transaction, error)

	// ListKeys returns the list of all the generated keys.
	ListKeys(walletName string) ([]GeneratedKey, error)

	// StartCoinJoin starts a CoinJoin round. It expects the wallet name, the password, a boolean to stop when all mixed and a boolean to override the pleb stop.
	StartCoinJoin(walletName string, password string, stopWhenAllMixed bool, overridePlebStop bool) error

	// StartCoinJoinSweep starts a CoinJoin to another wallet.
	StartCoinJoinSweep(walletName string, password string, outputWalletName string) error

	// StopCoinJoin stops a CoinJoin round.
	StopCoinJoin(walletName string) error

	// Stop stops and exits Wasabi.
	Stop() error

	// GetFeeRates returns the fee rates (in satoshi per byte) for the given confirmation targets (in blocks).
	GetFeeRates() (GetFeeRatesResponse, error)

	// ListWallets returns the list of all wallets.
	ListWallets() ([]ListWalletsResponseItem, error)

	// ExcludeFromCoinJoin excludes a coin from the CoinJoin or includes it again. It expects the wallet name, the transaction id and the index of the coin (vOut) and a boolean to exclude or include it.
	ExcludeFromCoinJoin(walletName string, txID string, index int, exclude bool) error

	// RecoverWallet recovers a wallet with the given name, mnemonic and password. The first parameter is the (new) wallet name, the second parameter is the mnemonic (recovery words), the third parameter is an optional passphrase (aka the password in Wasabi).
	RecoverWallet(walletName string, mnemonic string, password string) error

	// BuildUnsafeTransaction - constructs a transaction without checking fees and using unconfirmed coins. Unsafe, because no matter how big fee the user chooses, Wasabi will build the transaction. Potentially, the user can burn his money using this method, so be careful. The result is the transaction hex, waiting to be broadcast.
	BuildUnsafeTransaction(walletName string, payments []Payment, coins []Coin, feeTarget int, password string) (string, error)

	// PayInCoinJoin - pays to the specified address the specified amount of money using CoinJoin. Returns hte paymentId (UUID). A PayInCoinJoin is written to the logs of WasabiWallet, and it's status can be seen by using the ListPaymentsInCoinJoin method. Currently, the default maximum is 4 payments per client per CoinJoin. PayInCoinJoin only registers a payment, so if CoinJoin is not running or the amount is lower than the wallet balance, the payment is queued. Pending payments can be removed by using the CancelPaymentInCoinJoin method. Pending payments are also removed if the Wasabi client restarts.
	PayInCoinJoin(walletName string, address string, amount int, password string) (string, error)

	// ListPaymentsInCoinJoin - returns the list of payments in the CoinJoin.
	ListPaymentsInCoinJoin(walletName string) ([]ListPaymentsInCoinJoinResponseItem, error)

	// CancelPaymentInCoinJoin - cancels a payment in the CoinJoin. It expects the wallet name and the payment id.
	CancelPaymentInCoinJoin(walletName string, paymentID string) error

	// CancelTransaction - cancels a transaction and returns the transaction hex, ready for broadcast. It expects the wallet name, transaction id and the password. It is similar to the SpeedUpTransaction method, except that it will create a transaction back to the wallet. The transaction is not automatically broadcast.
	CancelTransaction(walletName string, txID string, password string) (string, error)

	// SpeedUpTransaction - speeds up a transaction and returns the transaction hex, ready for broadcast. It expects the wallet name, transaction id and the password. It does not automatically broadcast the new transaction, so it still needs to be (manually) broadcast.
	SpeedUpTransaction(walletName string, txID string, password string) (string, error)
}

// NewClient creates a new Client.
func NewClient(cfg Config) (Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	rpcClient := &client{
		host:    cfg.Host,
		port:    cfg.Port,
		headers: cfg.CustomHeaders,
	}
	if cfg.Transport == nil {
		rpcClient.httpClient = http.DefaultClient
	} else {
		rpcClient.httpClient = &http.Client{
			Transport: cfg.Transport,
		}
	}
	return rpcClient, nil
}

type client struct {
	httpClient *http.Client
	host       string
	port       int
	headers    map[string]string
	mutex      sync.Mutex
}

// Helper function
func (c *client) do(method Method, targetWalletName string, in, out interface{}) error {
	payload, err := encodeClientRequest(method.String(), in)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s:%d/%s", c.host, c.port, targetWalletName), bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	if c.headers != nil {
		for k, v := range c.headers {
			req.Header.Set(k, v)
		}
	}

	// Only one request at a time
	c.mutex.Lock()
	defer c.mutex.Unlock()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http status %v", resp.StatusCode)
	}
	defer resp.Body.Close()

	// Some methods return null, which is not an error. (LoadWallet, StopCoinJoin, Stop)
	if err := decodeClientResponse(resp.Body, out); err != nil && !errors.Is(err, RPCErrNullResult) {
		return err
	}
	return nil
}

// Method implementation

func (c *client) IsWasabiWalletUp() bool {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", c.host, c.port))
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

func (c *client) GetStatus() (resp GetStatusResponse, err error) {
	err = c.do(MethodGetStatus, "", nil, &resp)
	if err != nil {
		return GetStatusResponse{}, err
	}
	return
}

func (c *client) CreateWallet(walletName string, password string) (resp string, err error) {
	err = c.do(MethodCreateWallet, "", []interface{}{walletName, password}, &resp)
	if err != nil {
		return "", err
	}
	return
}

func (c *client) LoadWallet(walletName string) error {
	return c.do(MethodLoadWallet, "", []interface{}{walletName}, nil)
}

func (c *client) ListCoins(walletName string) (resp []ListCoinsResponse, err error) {
	err = c.do(MethodListCoins, walletName, nil, &resp)
	if err != nil {
		return nil, err
	}
	return
}

func (c *client) ListUnspentCoins(walletName string) (resp []ListCoinsResponse, err error) {
	err = c.do(MethodListUnspentCoins, walletName, nil, &resp)
	if err != nil {
		return nil, err
	}
	return
}

func (c *client) GetWalletInfo(walletName string) (resp GetWalletInfoResponse, err error) {
	err = c.do(MethodGetWalletInfo, walletName, nil, &resp)
	if err != nil {
		return GetWalletInfoResponse{}, err
	}
	return resp, nil
}

func (c *client) GetNewAddress(walletName string, label string) (resp GetNewAddressResponse, err error) {
	err = c.do(MethodGetNewAddress, walletName, []interface{}{label}, &resp)
	if err != nil {
		return GetNewAddressResponse{}, err
	}
	return resp, nil
}

func (c *client) Send(walletName string, payments []Payment, coins []Coin, feeTarget int, password string) (resp SendResponse, err error) {
	err = c.do(MethodSend, walletName, map[string]interface{}{"payments": payments, "coins": coins, "feeTarget": feeTarget, "password": password}, &resp)
	if err != nil {
		return SendResponse{}, err
	}
	return resp, nil
}

func (c *client) Build(walletName string, payments []Payment, coins []Coin, feeTarget int, password string) (resp string, err error) {
	err = c.do(MethodBuild, walletName, map[string]interface{}{"payments": payments, "coins": coins, "feeTarget": feeTarget, "password": password}, &resp)
	if err != nil {
		return "", err
	}
	return resp, nil
}

func (c *client) Broadcast(walletName string, hex string) (resp string, err error) {
	err = c.do(MethodBroadcast, walletName, []interface{}{hex}, &resp)
	if err != nil {
		return "", err
	}
	return resp, nil
}

func (c *client) GetHistory(walletName string) (resp []Transaction, err error) {
	err = c.do(MethodGetHistory, walletName, nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *client) ListKeys(walletName string) (resp []GeneratedKey, err error) {
	err = c.do(MethodListKeys, walletName, nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *client) StartCoinJoin(walletName string, password string, stopWhenAllMixed bool, overridePlebStop bool) error {
	return c.do(MethodStartCoinJoin, walletName, []interface{}{password, stopWhenAllMixed, overridePlebStop}, nil)
}

func (c *client) StartCoinJoinSweep(walletName string, password string, outputWalletName string) error {
	return c.do(MethodStartCoinJoinSweep, walletName, []interface{}{password, outputWalletName}, nil)
}

func (c *client) StopCoinJoin(walletName string) error {
	return c.do(MethodStopCoinJoin, walletName, nil, nil)
}

func (c *client) Stop() error {
	return c.do(MethodStop, "", nil, nil)
}

func (c *client) GetFeeRates() (resp GetFeeRatesResponse, err error) {
	err = c.do(MethodGetFeeRates, "", nil, &resp)
	if err != nil {
		return GetFeeRatesResponse{}, err
	}
	return resp, nil
}

func (c *client) ListWallets() (resp []ListWalletsResponseItem, err error) {
	err = c.do(MethodListWallets, "", nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *client) ExcludeFromCoinJoin(walletName string, txID string, index int, exclude bool) error {
	return c.do(MethodExcludeFromCoinJoin, walletName, []interface{}{txID, index, exclude}, nil)
}

func (c *client) RecoverWallet(walletName string, mnemonic string, password string) error {
	return c.do(MethodRecoverWallet, "", []interface{}{walletName, mnemonic, password}, nil)
}

func (c *client) BuildUnsafeTransaction(walletName string, payments []Payment, coins []Coin, feeTarget int, password string) (resp string, err error) {
	err = c.do(MethodBuildUnsafeTransaction, walletName, map[string]interface{}{"payments": payments, "coins": coins, "feeTarget": feeTarget, "password": password}, &resp)
	if err != nil {
		return "", err
	}
	return resp, nil
}

func (c *client) PayInCoinJoin(walletName string, address string, amount int, password string) (resp string, err error) {
	err = c.do(MethodPayInCoinJoin, walletName, []interface{}{address, amount, password}, &resp)
	if err != nil {
		return "", err
	}
	return resp, nil
}

func (c *client) ListPaymentsInCoinJoin(walletName string) (resp []ListPaymentsInCoinJoinResponseItem, err error) {
	err = c.do(MethodListPaymentsInCoinJoin, walletName, nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *client) CancelPaymentInCoinJoin(walletName string, paymentID string) error {
	return c.do(MethodCancelPaymentInCoinJoin, walletName, []interface{}{paymentID}, nil)
}

func (c *client) CancelTransaction(walletName string, txID string, password string) (resp string, err error) {
	err = c.do(MethodCancelTransaction, walletName, []interface{}{txID, password}, &resp)
	if err != nil {
		return "", err
	}
	return resp, nil
}

func (c *client) SpeedUpTransaction(walletName string, txID string, password string) (resp string, err error) {
	err = c.do(MethodSpeedUpTransaction, walletName, []interface{}{txID, password}, &resp)
	if err != nil {
		return "", err
	}
	return resp, nil
}

// encodeClientRequest encodes parameters for a JSON-RPC client request.
func encodeClientRequest(method string, args interface{}) ([]byte, error) {
	val, err := rand.Int(rand.Reader, big.NewInt(int64(math.MaxInt64)))
	if err != nil {
		return nil, fmt.Errorf("failed to generate request id: %w", err)
	}

	c := &clientRequest{
		Version: "2.0",
		Method:  method,
		Params:  args,
		Id:      val.Uint64(),
	}
	return json.Marshal(c)
}

// decodeClientResponse decodes the response body of a client request into the interface reply.
func decodeClientResponse(r io.Reader, reply interface{}) error {
	var c clientResponse
	if err := json.NewDecoder(r).Decode(&c); err != nil {
		return err
	}
	if c.Error != nil {
		jsonErr := &RPCError{}
		if err := json.Unmarshal(*c.Error, jsonErr); err != nil {
			return &RPCError{
				Code:    E_SERVER,
				Message: string(*c.Error),
			}
		}
		return jsonErr
	}

	if c.Result == nil {
		return RPCErrNullResult
	}

	return json.Unmarshal(*c.Result, reply)
}

// clientRequest represents a JSON-RPC request sent by a client.
type clientRequest struct {
	// JSON-RPC protocol.
	Version string `json:"jsonrpc"`

	// A String containing the name of the method to be invoked.
	Method string `json:"method"`

	// Object to pass as request parameter to the method.
	Params interface{} `json:"params"`

	// The request id. This can be of any type. It is used to match the
	// response with the request that it is replying to.
	Id uint64 `json:"id"`
}

// clientResponse represents a JSON-RPC response returned to a client.
type clientResponse struct {
	Version string           `json:"jsonrpc"`
	Result  *json.RawMessage `json:"result"`
	Error   *json.RawMessage `json:"error"`
}

type RPCErrorCode int

const (
	E_PARSE       RPCErrorCode = -32700
	E_INVALID_REQ RPCErrorCode = -32600
	E_NO_METHOD   RPCErrorCode = -32601
	E_BAD_PARAMS  RPCErrorCode = -32602
	E_INTERNAL    RPCErrorCode = -32603
	E_SERVER      RPCErrorCode = -32000
)

var RPCErrNullResult = errors.New("result is null")

type RPCError struct {
	// A Number that indicates the error type that occurred.
	Code RPCErrorCode `json:"code"` /* required */

	// A String providing a short description of the error.
	// The message SHOULD be limited to a concise single sentence.
	Message string `json:"message"` /* required */

	// A Primitive or Structured value that contains additional information about the error.
	Data interface{} `json:"data"` /* optional */
}

func (e *RPCError) Error() string {
	return e.Message
}
