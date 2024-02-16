package wasabi

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Config holds the configuration of a wasabi rpc client.
type Config struct {
	// Host is the address of the wasabi rpc server (without protocol and port).
	Host string
	// Port is the port of the wasabi rpc server. Default is 37128.
	Port int
	// CustomHeaders is a map of custom headers to send with the request
	CustomHeaders map[string]string
	// Transport is the http transport to use for the request. If nil, http.DefaultClient is used
	Transport http.RoundTripper
	// RpcUser is the rpc user to use for basic authentication
	RpcUser string
	// RpcPassword is the rpc password to use for basic authentication
	RpcPassword string
}

// Validate validates the config.
func (c *Config) Validate() error {
	switch {
	case c.Host == "":
		return fmt.Errorf("host must not be empty")
	case strings.ContainsAny(c.Host, "/:"):
		return fmt.Errorf("host must not contain / or :")
	case c.Port == 0:
		c.Port = 37128
	case c.Port < 0 || c.Port > 65535:
		return fmt.Errorf("port must be between 0 and 65535")
	case c.RpcUser != "" && c.RpcPassword == "":
		return fmt.Errorf("rpc password must not be empty if rpc user is set")
	case c.RpcUser == "" && c.RpcPassword != "":
		return fmt.Errorf("rpc user must not be empty if rpc password is set")
	case c.CustomHeaders != nil:
		for k, v := range c.CustomHeaders {
			if k == "" {
				return fmt.Errorf("custom header key must not be empty")
			}
			if v == "" {
				return fmt.Errorf("custom header value must not be empty")
			}
			if k == "Authorization" {
				return fmt.Errorf("custom header key must not be Authorization")
			}
		}
	case c.RpcUser != "" && c.RpcPassword != "":
		basicAuth := base64.StdEncoding.EncodeToString([]byte(c.RpcUser + ":" + c.RpcPassword))
		if c.CustomHeaders != nil {
			c.CustomHeaders["Authorization"] = "Basic " + basicAuth
		} else {
			c.CustomHeaders = map[string]string{
				"Authorization": "Basic " + basicAuth,
			}
		}
	}
	return nil
}

// BitcoinPeer provides information about a bitcoin peer.
type BitcoinPeer struct {
	IsConnected bool      `json:"isConnected"`
	LastSeen    time.Time `json:"lastSeen"`
	Endpoint    string    `json:"endpoint"`
	UserAgent   string    `json:"userAgent"`
}

// GetStatusResponse provides the response of a getstatus request.
type GetStatusResponse struct {
	TorStatus            TorStatus      `json:"torStatus"`
	BackendStatus        BackendStatus  `json:"backendStatus"`
	BestBlockchainHeight uint64         `json:"bestBlockchainHeight,string"`
	BestBlockchainHash   string         `json:"bestBlockchainHash"`
	FiltersCount         int            `json:"filtersCount"`
	FiltersLeft          int            `json:"filtersLeft"`
	Network              BitcoinNetwork `json:"network"`
	ExchangeRate         float64        `json:"exchangeRate"`
	Peers                []BitcoinPeer  `json:"peers"`
}

// ListCoinsResponse provides the response of a listcoins request.
type ListCoinsResponse struct {
	TxID                 string  `json:"txid"`
	Index                int     `json:"index"`
	Amount               int     `json:"amount"`
	AnonymityScore       float64 `json:"anonymityScore"`
	Confirmed            bool    `json:"confirmed"`
	Confirmations        int     `json:"confirmations"`
	KeyPath              string  `json:"keyPath"`
	Address              string  `json:"address"`
	SpentBy              *string `json:"spentBy,omitempty"` // may be null
	Label                string  `json:"label,omitempty"`
	ExcludedFromCoinJoin bool    `json:"excludedFromCoinjoin"`
}

// GetWalletInfoResponse provides the response of a getwalletinfo request.
type GetWalletInfoResponse struct {
	WalletName           string              `json:"walletName"`
	WalletFile           string              `json:"walletFile"`
	State                WalletState         `json:"state"`
	MasterKeyFingerprint string              `json:"masterKeyFingerprint"`
	AnonScoreTarget      int                 `json:"anonScoreTarget"`
	IsWatchOnly          bool                `json:"isWatchOnly"`
	IsHardwareWallet     bool                `json:"isHardwareWallet"`
	IsAutoCoinJoin       bool                `json:"isAutoCoinjoin"`
	IsRedCoinIsolation   bool                `json:"isRedCoinIsolation"`
	Accounts             []WalletInfoAccount `json:"accounts"`
	Balance              int                 `json:"balance,omitempty"`
	CoinJoinStatus       CoinJoinStatus      `json:"coinjoinStatus,omitempty"`
}

// WalletInfoAccount provides information about a wallet account.
type WalletInfoAccount struct {
	Name      string `json:"name"`
	PublicKey string `json:"publicKey"`
	KeyPath   string `json:"keyPath"`
}

// GetNewAddressResponse provides the response of a getnewaddress request.
type GetNewAddressResponse struct {
	Address      string `json:"address"`
	KeyPath      string `json:"keyPath"`
	Label        string `json:"label"`
	PublicKey    string `json:"publicKey"`
	ScriptPubKey string `json:"scriptPubKey"`
}

// SendResponse provides the response of a send request.
type SendResponse struct {
	TransactionID string `json:"txid"`
	Transaction   string `json:"tx"`
}

// Payment provides information about a payment.
type Payment struct { // PaymentInfo
	SendTo string `json:"sendto"`
	Amount int    `json:"amount"`
	Label  string `json:"label"`
}

// Coin provides information about a coin.
type Coin struct { // OutPoint
	TransactionID string `json:"transactionid"`
	Index         int    `json:"index"`
}

// Transaction provides information about a transaction in history.
type Transaction struct {
	DateTime         time.Time `json:"datetime"`
	Height           int       `json:"height"`
	Amount           int       `json:"amount"`
	Label            string    `json:"label"`
	Tx               string    `json:"tx"`
	IsLikelyCoinJoin bool      `json:"islikelycoinjoin"`
}

// GeneratedKey provides information about a generated key.
type GeneratedKey struct {
	FullKeyPath  string `json:"fullKeyPath"`
	Internal     bool   `json:"internal"`
	KeyState     int    `json:"keyState"`
	Label        string `json:"label"`
	ScriptPubKey string `json:"scriptPubKey"`
	PubKey       string `json:"pubkey"`
	PubKeyHash   string `json:"pubKeyHash"`
	Address      string `json:"address"`
}

// GetFeeRatesResponse provides the response of a getfeerates request. It is a map of confirmation target (in blocks) to fee rate (in satoshi per byte).
type GetFeeRatesResponse map[string]int

// ListWalletsResponseItem provides the response of a listwallets request.
type ListWalletsResponseItem struct {
	Name string `json:"walletName"`
}

// ListPaymentsInCoinJoinResponseItem provides the item of a listpaymentsincoinjoin response list.
type ListPaymentsInCoinJoinResponseItem struct {
	// ID is the id of the payment (UUID). That id can be used to cancel the payment.
	ID string `json:"id"`
	// Amount is the amount of the payment in satoshi.
	Amount int `json:"amount"`
	// Destination is the destination of the payment (ScriptPubKey hex).
	Destination string `json:"destination"`
	// State is the state history of the payment.
	State []PaymentInCoinJoinStateHistoryItem `json:"state"`
	// Address is the address of the payment.
	Address string `json:"address"`
}

// PaymentInCoinJoinStateHistoryItem provides the item of a payment in coinjoin state history list.
type PaymentInCoinJoinStateHistoryItem struct {
	Status PaymentStatus `json:"status"`
	Round  int           `json:"round,omitempty"`
	TxID   string        `json:"txid,omitempty"`
}
