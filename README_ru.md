<figure align="center" style="height:128px;">
    <img src="https://github.com/acfnv/go-wasabi-rpc-client/raw/master/media/images/Alexey_Navalny.png" alt="Фото Алексея Навального" height="128" align="left"/>
	<blockquote style="padding-bottom: 24px;">
		<h3><i>«Все, что нужно для торжества зла, — это бездействие добрых людей»</i></h3>
	</blockquote>
	<figcaption><i> – <a href="https://acf.international/ru/faces" target="_blank">Алексей Навальный</a>, Российский оппозиционный политик, основатель Фонда Борьбы с Коррупцией, последовательный противник Владимира Путина и его режима.</i></figcaption>
</figure>

---
Go Wasabi RPC Client
====================
[![GoDoc](https://godoc.org/github.com/acfnv/go-wasabi-rpc-client?status.svg)](https://godoc.org/github.com/acfnv/go-wasabi-rpc-client)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/acfnv/go-wasabi-rpc-client/raw/master/LICENSE)

<p align="center">
<img src="https://github.com/acfnv/go-wasabi-rpc-client/raw/master/media/images/go-wasabi-logo.png" alt="Wasabi Gopher" width="320" />
</p>

# Реализация клиента для взаимодействия с WasabiWallet через RPC на языке программирования Go.

Вы можете использовать этот клиент для взаимодействия с WasabiWallet через RPC.
Текущая редакция клиента поддерживает все методы, которые используются в WasabiWallet в основной ветке репозитория по состоянию на 16.02.2024 года. Это означает что некоторые методы еще не доступны в текущем релизе кошелька и отчасти отсутствуют в документации к RPC на сайте WasabiWallet.

Клиент можно использовать с кошельком WasabiWallet, как запущенным в режиме GUI, так и в режиме headless daemon.

## Ссылки на ресурсы WasabiWallet:

- [Веб-сайт](https://wasabiwallet.io/)
- [Репозиторий кошелька на GitHub](https://github.com/zkSNACKs/WalletWasabi)
- [Репозиторий документации¹](https://github.com/zkSNACKs/WasabiDoc)
- [Документация к RPC²](https://docs.wasabiwallet.io/using-wasabi/RPC.html)

<i>
¹ - Документация к WasabiWallet на GitHub, содержит документацию ко всем методам RPC, в том числе и тем, которые еще не доступны в релизе кошелька. Такие методы описаны в правках, которые находятся в разделе пулл-реквестов. Документацию к таким методам вы можете найти через поиск по репозиторию.
² - Документация к RPC на сайте WasabiWallet, содержит документацию только к методам, которые доступны в текущем релизе кошелька.
</i>

## Установка WasabiWallet.
### Вариант 1: Установка кошелька из текущего релиза:

- [Ссылка для загрузки кошелька с GUI](https://wasabiwallet.io/#download)
- [Инструкция по запуску кошелька в режиме headless daemon](https://docs.wasabiwallet.io/using-wasabi/Daemon.html#introduction)

### Вариант 2: Сборка кошелька из исходников и запуск в режиме демона в Docker-контейнере на базе debian:bookworm-20240211:

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
Подробная информация о параметрах запуска кошелька доступна по [ссылке](https://docs.wasabiwallet.io/using-wasabi/StartupParameters.html#config-file-configurations).

#### Сборка образа и запуск контейнера с кошельком:

```bash
docker build --target wallet .

docker run -d --name wasabi-wallet -p 37128:37128 wallet

```

## Использование нашего клиента.

### Загрузка клиента:

```bash
go get -u github.com/acfnv/go-wasabi-rpc-client
```

### Пример использования:

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
	
	// Определяем конфигурацию для подключения к WasabiWallet.
	config := wasabi.Config{
		Host: *rpcHost,
		Port: *rpcPort,
	}
	
	// Создаем клиент для взаимодействия с WasabiWallet.
	client, err := wasabi.NewClient(config)
	if err != nil {
		return nil, err
	}
	
	// Ожидаем запуска WasabiWallet.
	for {
		if client.IsWasabiWalletUp() {
			break
		}
		time.Sleep(1 * time.Second)
	}
	log.Printf("WasabiWallet RPC Server started")
	
	// Выполнение запроса к WasabiWallet.
	resp, err := client.GetStatus()
	if err != nil {
		return err
	}
	log.Printf("WasabiWallet RPC Server status: %v", resp)
	
	// ...
}

```

## Вклад в разработку
* Вы можете форкнуть этот репозиторий, дополнить его и внести свой вклад.
* Вы можете внести свой вклад через пулл-реквесты.

## Поддержать новые разработки:

Вы можете [поддержать ФБК](https://donate.acf.international/en), используя банковские карты (если Вы находитесь за пределами РФ), электронные кошельки и криптовалюты.

## Лицензия
[Свободная лицензия MIT](https://github.com/acfnv/go-wasabi-rpc-client/raw/master/LICENSE)
