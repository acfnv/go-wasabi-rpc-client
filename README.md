<figure align="center" style="height:128px;">
    <img src="https://github.com/acfnv/go-wasabi-rpc-client/raw/master/media/images/Alexey_Navalny.png" alt="Photo of Alexey Navalny" height="128" align="left" />
	<blockquote style="padding-bottom: 24px;">
		<h3><i>“The only thing necessary for the triumph of evil is for good men to do nothing.”</i></h3>
	</blockquote>
	<figcaption><i> – <a href="https://acf.international/ru/faces" target="_blank">Alexey Navalny</a>, Russian opposition politician, founder of the Anti-Corruption Foundation, consistent opponent of Vladimir Putin and his regime.</i></figcaption>
</figure>

---
Go Wasabi RPC Client
====================
[![GoDoc](https://godoc.org/github.com/acfnv/go-wasabi-rpc-client?status.svg)](https://godoc.org/github.com/acfnv/go-wasabi-rpc-client)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/acfnv/go-wasabi-rpc-client/raw/master/LICENSE)

<p align="center">
<img src="https://github.com/acfnv/go-wasabi-rpc-client/raw/master/media/images/go-wasabi-logo.png" alt="Wasabi Gopher" width="320" />
</p>

# Implementation of a client for interacting with WasabiWallet via RPC in the Go programming language.

You can use this client to interact with WasabiWallet via RPC.
The current version of the client supports all methods used in WasabiWallet in the main repository branch as of February 16, 2024. This means that some methods are not yet available in the current wallet release and are partially absent from the RPC documentation on the WasabiWallet website.

The client can be used with WasabiWallet, both when running in GUI mode and as a headless daemon.

## Links to WasabiWallet resources:

- [Website](https://wasabiwallet.io/)
- [Wallet repository on GitHub](https://github.com/zkSNACKs/WalletWasabi)
- [Documentation repository¹](https://github.com/zkSNACKs/WasabiDoc)
- [RPC documentation²](https://docs.wasabiwallet.io/using-wasabi/RPC.html)

<i>
¹ - WasabiWallet documentation on GitHub contains documentation for all RPC methods, including those not yet available in the wallet release. Such methods are described in the pull requests section. You can find documentation for these methods by searching the repository.
² - RPC documentation on the WasabiWallet website contains documentation only for methods available in the current wallet release.
</i>

## Installing WasabiWallet.
### Option 1: Installing the wallet from the current release:

- [Link to download the GUI wallet](https://wasabiwallet.io/#download)
- [Instructions for running the wallet as a headless daemon](https://docs.wasabiwallet.io/using-wasabi/Daemon.html#introduction)

### Option 2: Building the wallet from source and running as a headless daemon in a Docker container under debian:bookworm-20240211:

```Dockerfile
# Dockerfile for wasabi-wallet

FROM debian:bookworm-20240211 AS wallet

ENV COMMIT_HASH 848df269f1021bacc5d7d85e683ef4384a87cde6

RUN apt-get update && apt-get install -y \
    wget \
    software-properties-common \
    netcat-openbsd \
    curl \
    gnupg2 \
    git \
    libevent-2.1-7 && \
    wget https://packages.microsoft.com/config/debian/12/packages-microsoft-prod.deb -O packages-microsoft-prod.deb && \
    dpkg -i packages-microsoft-prod.deb && \
    rm packages-microsoft-prod.deb && \
    apt-get update && apt-get install -y \
    apt-transport-https \
    dotnet-sdk-8.0 && \
    apt-get clean && rm -rf /var/lib/apt/lists/* /var/cache/apt/*

RUN mkdir -p /wasabi/data

ENV DOTNET_CLI_TELEMETRY_OPTOUT=1

RUN git clone --single-branch --branch=master https://github.com/zkSNACKs/WalletWasabi.git \
    && cd WalletWasabi \
    && git checkout $COMMIT_HASH 

RUN cd /WalletWasabi/WalletWasabi.Fluent.Desktop && dotnet build && rm -rf ~/.nuget ~/.local

WORKDIR /WalletWasabi/WalletWasabi.Daemon

CMD ["dotnet", "run", "--jsonrpcserverenabled=true", "--enablegpu=false", "--datadir=\"/wasabi/data\"", "--network=testnet"]

```
Detailed information about wallet startup parameters is available [here](https://docs.wasabiwallet.io/using-wasabi/StartupParameters.html#config-file-configurations).

#### Building the image and running the container with the wallet:

```bash
docker build --target wallet -t wasabi_wallet .

docker run -d -p 37128:37128/tcp --name wasabi_container wasabi_wallet

```

#### Check the wallet logs:

```bash
docker logs -f wasabi_container
```

## Using our client.

### Downloading the client:

```bash
go get -u github.com/acfnv/go-wasabi-rpc-client
```

### Example of usage:

```go
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
	
	"github.com/acfnv/go-wasabi-rpc-client"
)

var (
	rpcHost     = flag.String("rpc_host", "127.0.0.1", "Host of the wasabi rpc server.")
	rpcPort     = flag.Int("rpc_port", 37128, "Port of the wasabi rpc server.")
)

func main() {
	flag.Parse()
	
	// Define the configuration for the RPC client.
	config := wasabi.Config{
		Host: *rpcHost,
		Port: *rpcPort,
	}
	
	// Create a new RPC client.
	client, err := wasabi.NewClient(config)
	if err != nil {
		return nil, err
	}
	
	// Check the availability of the RPC server.
	for {
		if client.IsWasabiWalletUp() {
			break
		}
		time.Sleep(1 * time.Second)
	}
	log.Printf("WasabiWallet RPC Server started")
	
	// Making a request to the RPC server.
	resp, err := client.GetStatus()
	if err != nil {
		return err
	}
	log.Printf("WasabiWallet RPC Server status: %v", resp)
	
	// ...
}

```

## Contribution
* You can fork this, extend it and contribute back.
* You can contribute with pull requests.

## Support further developments:

You can [support the Anti-Corruption Foundation (ACF)](https://donate.acf.international/en), using bank cards, electronic wallets, and cryptocurrencies.

## LICENSE
[MIT License](https://github.com/acfnv/go-wasabi-rpc-client/raw/master/LICENSE)
