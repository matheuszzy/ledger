# Ledger

A small in-memory account API built for the EBANX Software Engineer take-home. It supports deposits, withdrawals and transfers between accounts.

## Requirements

- Go 1.25+

## Run

```sh
go run ./cmd/api
```

The server listens on port `8080`. Set `PORT` to use another port:

```sh
PORT=3000 go run ./cmd/api
```

To build a binary instead:

```sh
go build -o ledger ./cmd/api
./ledger
```

When deploying, set `GIN_MODE=release` to turn off Gin's debug logs.

## Test

```sh
go test -race ./...
```

| Level | Location | What it covers |
|---|---|---|
| Unit | `internal/account/domain` | Balance rules: deposit, withdraw, negative amounts, insufficient funds |
| Unit | `internal/account/repository` | Save and load, stored copies are isolated, concurrent access |
| Service | `internal/account/service` | Each operation against the real in-memory store: state changes, error cases, concurrent deposits, withdrawals and transfers |
| Integration | `internal/account/handler/handler_integration_test.go` | Real HTTP server: the full spec scenario, end-to-end workflows, error responses, no side effects on `GET` |

No mocks are used. Every test runs against the real in-memory repository and checks the resulting state, not only the response.

Run only the integration tests:

```sh
go test -race ./internal/account/handler/
```

## API

| Method | Path | Description |
|---|---|---|
| `POST` | `/reset` | Removes every account |
| `GET` | `/balance?account_id={id}` | Returns the balance as plain text |
| `POST` | `/event` | `deposit`, `withdraw` or `transfer` |

Example session, with the server running on port `8080`:

```sh
curl -X POST localhost:8080/reset
# 200 OK

curl -X POST localhost:8080/event -d '{"type":"deposit", "destination":"100", "amount":10}'
# 201 {"destination":{"id":"100","balance":10}}

curl -X POST localhost:8080/event -d '{"type":"withdraw", "origin":"100", "amount":5}'
# 201 {"origin":{"id":"100","balance":5}}

curl -X POST localhost:8080/event -d '{"type":"transfer", "origin":"100", "amount":5, "destination":"300"}'
# 201 {"origin":{"id":"100","balance":0},"destination":{"id":"300","balance":5}}

curl "localhost:8080/balance?account_id=300"
# 200 5
```

### Responses

| Status | Body | When |
|---|---|---|
| `200` | balance or `OK` | `GET /balance` on an existing account, `POST /reset` |
| `201` | affected accounts as JSON | Successful event |
| `404` | `0` | Unknown account on `GET /balance`, or unknown origin on withdraw or transfer |
| `422` | `0` | Insufficient funds on withdraw or transfer |
| `400` | `CODE - message` | Invalid input: malformed JSON, unknown event type, missing ID, negative amount, transfer to the same account |
| `405` | `405 method not allowed` | Wrong HTTP method on a known route |

The `404` and `422` bodies follow the spec's `0` convention. Validation errors carry a code (`TE` for transport, `AC` for account rules) so clients can tell them apart without parsing the message.

## Project layout

```
cmd/api/                   entrypoint: wires repository -> service -> handler
internal/account/
  domain/                  Account and Balance entities, business rules, domain errors
  repository/              in-memory store
  service/                 use cases: Deposit, Withdraw, Transfer, Balance, Reset
  handler/                 HTTP: decode request, call service, encode response
```

Dependencies only point inward:

- `handler` depends on `service`.
- `service` depends on `domain` and on a small `Repository` interface that it defines itself.
- `domain` depends on nothing.

The handler holds no business logic, so the HTTP layer can be replaced without touching the rules.

## Decisions

- **In-memory storage.** Durability is not required, so accounts live in a map guarded by a mutex. To add a database, implement the three-method `Repository` interface in `internal/account/service/service.go`.
- **The repository stores copies.** `GetByID` and `Save` copy the account, so stored state only changes when the service calls `Save`. A failed operation cannot leave a half-updated account behind.
- **Atomic events.** One mutex in the service covers each deposit, withdrawal and transfer. A transfer validates both sides first and saves them only if the withdrawal and the deposit both succeed. Concurrent tests run with `-race` check that money is never lost or created.
- **Reads have no side effects.** `GET /balance` never creates an account, and a test covers it.
- **Account creation.** Deposits and transfers create the destination account if it does not exist. Withdrawals and transfers from an unknown origin return `404`, as the spec requires.
- **Insufficient funds returns** `422`**.** The spec does not define this case. `422` keeps it distinct from an unknown account (`404`) and from malformed input (`400`).
- **Gin** handles routing, JSON binding and `405` responses, which keeps the handler short.

## Spec

The scenario from the assignment:

```
--
# Reset state before starting tests
POST /reset
200 OK

--
# Get balance for non-existing account
GET /balance?account_id=1234
404 0

--
# Create account with initial balance
POST /event {"type":"deposit", "destination":"100", "amount":10}
201 {"destination": {"id":"100", "balance":10}}

--
# Deposit into existing account
POST /event {"type":"deposit", "destination":"100", "amount":10}
201 {"destination": {"id":"100", "balance":20}}

--
# Get balance for existing account
GET /balance?account_id=100
200 20

--
# Withdraw from non-existing account
POST /event {"type":"withdraw", "origin":"200", "amount":10}
404 0

--
# Withdraw from existing account
POST /event {"type":"withdraw", "origin":"100", "amount":5}
201 {"origin": {"id":"100", "balance":15}}

--
# Transfer from existing account
POST /event {"type":"transfer", "origin":"100", "amount":15, "destination":"300"}
201 {"origin": {"id":"100", "balance":0}, "destination": {"id":"300", "balance":15}}

--
# Transfer from non-existing account
POST /event {"type":"transfer", "origin":"200", "amount":15, "destination":"300"}
404 0
```
