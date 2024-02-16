package wasabi

// Method is a wasabi RPC method.
type Method string

const (
	MethodGetStatus               Method = "getstatus"
	MethodCreateWallet            Method = "createwallet"
	MethodLoadWallet              Method = "loadwallet"
	MethodListCoins               Method = "listcoins"
	MethodListUnspentCoins        Method = "listunspentcoins"
	MethodGetWalletInfo           Method = "getwalletinfo"
	MethodGetNewAddress           Method = "getnewaddress"
	MethodSend                    Method = "send"
	MethodBuild                   Method = "build"
	MethodBroadcast               Method = "broadcast"
	MethodGetHistory              Method = "gethistory"
	MethodListKeys                Method = "listkeys"
	MethodStartCoinJoin           Method = "startcoinjoin"
	MethodStartCoinJoinSweep      Method = "startcoinjoinsweep"
	MethodStopCoinJoin            Method = "stopcoinjoin"
	MethodStop                    Method = "stop"
	MethodGetFeeRates             Method = "getfeerates"
	MethodListWallets             Method = "listwallets"
	MethodExcludeFromCoinJoin     Method = "excludefromcoinjoin"
	MethodRecoverWallet           Method = "recoverwallet"
	MethodBuildUnsafeTransaction  Method = "buildunsafetransaction"
	MethodPayInCoinJoin           Method = "payincoinjoin"
	MethodListPaymentsInCoinJoin  Method = "listpaymentsincoinjoin"
	MethodCancelPaymentInCoinJoin Method = "cancelpaymentincoinjoin"
	MethodCancelTransaction       Method = "canceltransaction"
	MethodSpeedUpTransaction      Method = "speeduptransaction"
)

// String returns the string representation of the method.
func (m Method) String() string {
	return string(m)
}

// BitcoinNetwork is a bitcoin network.
type BitcoinNetwork string

const (
	BitcoinNetworkMainnet BitcoinNetwork = "Main"
	BitcoinNetworkTestnet BitcoinNetwork = "TestNet"
)

// CoinJoinStatus is a coinjoin status.
type CoinJoinStatus string

const (
	CoinJoinStatusIdle            CoinJoinStatus = "Idle"
	CoinJoinStatusInSchedule      CoinJoinStatus = "In schedule"
	CoinJoinStatusInProgress      CoinJoinStatus = "In progress"
	CoinJoinStatusInCriticalPhase CoinJoinStatus = "In critical phase"
)

// PaymentStatus is a payment status.
type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "Pending"
	PaymentStatusInProgress PaymentStatus = "In progress"
	PaymentStatusFinished   PaymentStatus = "Finished"
)

// BackendStatus is a status of wasabi backend.
type BackendStatus string

const (
	BackendStatusConnected    BackendStatus = "Connected"
	BackendStatusDisconnected BackendStatus = "Disconnected"
)

// TorStatus is a status of tor.
type TorStatus string

const (
	TorStatusNotRunning TorStatus = "Not running"
	TorStatusRunning    TorStatus = "Running"
	TorStatusTurnedOff  TorStatus = "Turned off"
)

// WalletState is a state of wallet.
type WalletState string

const (
	WalletStateUninitialized  WalletState = "Uninitialized"
	WalletStateWaitingForInit WalletState = "WaitingForInit"
	WalletStateInitialized    WalletState = "Initialized"
	WalletStateStarting       WalletState = "Starting"
	WalletStateStarted        WalletState = "Started"
	WalletStateStopping       WalletState = "Stopping"
	WalletStateStopped        WalletState = "Stopped"
)

// WalletError is a wallet error.
type WalletError string

const (
	ErrorWalletIsNotFullyLoadedYet        WalletError = "Wallet is not fully loaded yet."
	ErrorIndexFileInconsistency           WalletError = "Index file inconsistency detected."
	ErrorNegativeIssuerBalance            WalletError = "Negative issuer balance"
	ErrorNegativeBalance                  WalletError = "Negative balance"
	ErrorIncorrectPassword                WalletError = "Incorrect password."
	ErrorPaymentNotPending                WalletError = "Payment could not be canceled because it is not pending."
	ErrorPaymentNotFound                  WalletError = "Payment was not found."
	ErrorNotEnoughCoins                   WalletError = "Not enough coins registered to participate in the coinjoin."
	ErrorNoSecretInTheWatchOnlyMode       WalletError = "No secret in the watch-only mode."
	ErrorOutputWalletNameInvalid          WalletError = "Output wallet name is invalid."
	ErrorRPCMethodSpecial                 WalletError = "This RPC method is special and the handling method should not be called."
	ErrorCoinJoinResultTypeNotHandled     WalletError = "The coinjoin result type was not handled."
	ErrorBlameRoundsNotSuccessful         WalletError = "Blame rounds were not successful."
	ErrorNotPossibleToSubtractTheFee      WalletError = "Not possible to subtract the fee."
	ErrorOriginalPSBTShouldNotBeFinalized WalletError = "The original PSBT should not be finalized."
	ErrorTransactionNotCancellable        WalletError = "Transaction is not cancellable."
	ErrorTransactionNotSpeedupable        WalletError = "Transaction is not speedupable."
	ErrorCannotGetFeeEstimations          WalletError = "Cannot get fee estimations."
)

func (e WalletError) Error() string {
	return string(e)
}
