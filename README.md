# stress-test

CLI em Go para realizar **testes de carga em serviços web**. Você informa a URL, o número total de requisições e o nível de concorrência, e ao final recebe um relatório com o tempo total, a distribuição dos status HTTP e o tempo médio de resposta de cada status.

## O que o projeto faz

- Dispara requisições `GET` para a URL informada.
- Distribui o trabalho entre N workers (goroutines), conforme a concorrência definida.
- Garante que o número total de requisições seja **exatamente** o informado em `--requests`.
- Mede o tempo de cada requisição (até a leitura completa do body) e calcula a média por status.
- Exibe um relatório no console ao final da execução.

### Exemplo de saída

```
Iniciando teste de carga: url=http://host.docker.internal:8080/status/200,404,500,200,200 requests=10000 concurrency=10

========== Relatório do Teste de Carga ==========
Tempo total gasto:            2.314s
Total de requests realizados: 10000
Requests com status 200:      6012 (tempo médio: 2.287ms)
Distribuição de outros status:
  404: 1990 (tempo médio: 2.301ms)
  500: 1998 (tempo médio: 2.275ms)
=================================================
```

Requisições que falham sem resposta HTTP (timeout, DNS, conexão recusada) aparecem em uma linha separada, junto com o tempo médio até a falha.

## Parâmetros

| Flag            | Descrição                                   | Obrigatório |
|-----------------|---------------------------------------------|-------------|
| `--url`         | URL do serviço a ser testado (http ou https) | Sim         |
| `--requests`    | Número total de requisições (> 0)            | Sim         |
| `--concurrency` | Número de chamadas simultâneas (> 0)         | Sim         |

Se `--concurrency` for maior que `--requests`, a concorrência é reduzida para o valor de `--requests`.

## Estrutura do projeto

```
.
├── cmd/
│   └── main.go              # flags, validação e orquestração
├── internal/
│   ├── runner/
│   │   └── runner.go        # execução do teste (worker pool) e coleta de métricas
│   └── report/
│       └── report.go        # formatação do relatório
├── Dockerfile
└── go.mod
```

## Como buildar e executar

### Com Docker (recomendado)

**Build da imagem:**

```bash
docker build -t stress-test .
```

**Execução:**

```bash
docker run stress-test --url=http://google.com --requests=1000 --concurrency=10
```

### Sem Docker

Requer Go instalado (a versão usada no `Dockerfile` é a 1.26).

**Rodando direto:**

```bash
go run ./cmd --url=http://google.com --requests=1000 --concurrency=10
```

**Gerando o binário:**

```bash
go build -o stress-test ./cmd
./stress-test --url=http://google.com --requests=1000 --concurrency=10
```

## Testando contra um serviço local

Disparar muitas requisições contra serviços públicos pode causar bloqueios e não é uma boa prática. Para testar, suba um serviço local, como o [go-httpbin](https://github.com/mccutchen/go-httpbin), cujo endpoint `/status/{códigos}` sorteia entre os status informados:

```bash
docker run --rm -p 8080:8080 mccutchen/go-httpbin
```

### Atenção ao `localhost` dentro do container

Dentro do container, `localhost` aponta para o próprio container, não para a sua máquina. Use uma das opções abaixo.

**Mac/Windows (Docker Desktop):**

```bash
docker run stress-test \
  --url=http://host.docker.internal:8080/status/200,404,500,200,200 \
  --requests=10000 --concurrency=10
```

**Linux (rede do host):**

```bash
docker run --network host stress-test \
  --url=http://localhost:8080/status/200,404,500,200,200 \
  --requests=10000 --concurrency=10
```

**Qualquer SO (rede Docker compartilhada):**

```bash
docker network create loadnet

docker run -d --rm --name httpbin --network loadnet mccutchen/go-httpbin

docker run --network loadnet stress-test \
  --url=http://httpbin:8080/status/200,404,500,200,200 \
  --requests=10000 --concurrency=10
```

## Comportamento e limitações

- Apenas o método `GET` é utilizado.
- O timeout de cada requisição é de 30 segundos.
- Redirecionamentos são seguidos automaticamente, então o status contabilizado é o da resposta final (por exemplo, `http://google.com` tende a aparecer como 200, e não como 301).
- O `Dockerfile` compila para `linux/amd64`. Em máquinas ARM (como Macs com Apple Silicon), a imagem roda sob emulação, o que pode afetar o desempenho do teste.