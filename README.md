# Digital Wallet Service

A high-concurrency Digital Wallet service written in Go, designed to handle financial transactions securely and efficiently. This service implements a **Service-Repository** architecture and ensures data consistency using **Database Transactions** and **Row Locking** to prevent race conditions and deadlocks.

## 🚀 Features

*   **Wallet Management**: Create wallets and retrieve balances using the **Cache-Aside** pattern (Redis).
*   **Secure Transfers**: Concurrency-safe money transfers between wallets using `SELECT FOR UPDATE` and deterministic lock ordering.
*   **Event-Driven Architecture**: Asynchronous processing of transfer events using **RabbitMQ** (e.g., email notifications).
*   **Audit Logging**: All transactions are recorded in a permanent ledger.
*   **Graceful Shutdown**: Handles OS signals to cleanly close connections and stop servers.

## 🛠️ Technology Stack

*   **Language**: [Golang 1.22+](https://go.dev/) (Standard Library `net/http` with new routing)
*   **Database**: [PostgreSQL](https://www.postgresql.org/) (Driver: `pgx`, ORM: `GORM`)
*   **Caching**: [Redis](https://redis.io/) (`go-redis/v9`)
*   **Messaging**: [RabbitMQ](https://www.rabbitmq.com/) (`amqp091-go`)
*   **Validation**: `go-playground/validator`
*   **Testing**: Concurrency integration tests.

## ⚙️ Prerequisites

Ensure you have the following running in your environment:

1.  **Go** (v1.22 or higher)
2.  **PostgreSQL** (Default: `localhost:5432`)
3.  **Redis** (Default: `localhost:6379`)
4.  **RabbitMQ** (Default: `localhost:5672`)

> **Tip:** You can use Docker to quickly spin up infrastructure:
> ```bash
> docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=wallet_db postgres
> docker run -d -p 6379:6379 redis
> docker run -d -p 5672:5672 -p 15672:15672 rabbitmq:management
> ```

## 🏃 How to Run

1.  **Clone the repository**
    ```bash
    git clone git@github.com:msfalif404/high-concurrency-digital-wallet-with-golang.git
    cd digital-wallet
    ```

2.  **Install Dependencies**
    ```bash
    go mod tidy
    ```

3.  **Configure Environment** (Optional)
    Set environment variables if your services are on different ports:
    ```bash
    export DATABASE_URL="host=localhost user=postgres password=postgres dbname=wallet_db port=5432 sslmode=disable"
    export REDIS_ADDR="localhost:6379"
    export RABBITMQ_URL="amqp://guest:guest@localhost:5672/"
    ```

4.  **Run the Server**
    ```bash
    go run cmd/server/main.go
    ```
    The server will start on `http://localhost:8080`.

## 🧪 Running Tests

To verify the concurrency safety (race condition checks), run the integration test:

```bash
go test -v ./tests/...
```

This test spawns **50 concurrent goroutines** to transfer money between wallets and asserts that the final balance matches exactly.

## 📡 API Endpoints

### 1. Create Wallet
**POST** `/wallets`
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### 2. Get Balance
**GET** `/wallets/{id}`

### 3. Transfer Money
**POST** `/transfers`
```json
{
  "sender_id": "sender-uuid",
  "receiver_id": "receiver-uuid",
  "amount": 100
}
```
*Amount is in cents (e.g., 100 = $1.00).*
