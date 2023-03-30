# Backend Internship Test Task 🤓

> Microservice for managing user balances

### Problem 😤

Our company has many microservices. Many of them somehow need to interact with user balance. The architectural committee has decided to centralize the work with user balance in a separate service.

### Task 🫣

Implement a microservice for working with user balances (credit, debit, transfer, and getting user balance). The service should provide an HTTP API with requests and responses in JSON format.

Main task (minimum): 👍

1. Method for crediting funds to the balance. Accepts user ID and the amount of funds to credit.
2. Method for reserving funds from the main balance to a separate account. Accepts user ID, service ID, order ID, and cost. 
3. Method for recognizing revenue. Deducts money from the reserve and adds data to the accounting report. Accepts user ID, service ID, order ID, and amount. 
4. Method for getting the user balance. Accepts user ID.

Optional: ✅

1. Code coverage with tests. 
2. Swagger file for your API. 
3. Implementation of a scenario for unreserving funds if the service cannot be applied.

Full task [text](./docs/task.md), swagger.yaml [here](./api/swagger.yaml)

## Installation 🌚
First of all, clone this repo and download dependencies. **Requirements**: `go  1.20`, `git`

```shell
git clone git@github.com:pershin-daniil/internship_backend_2022.git
cd internship_backend_2022
go mod tidy
```

`make up` - to start docker container, and then you need databases. So, [here](./pkg/pgstore/create_table.sql) file to create tables.

Run command to start server. This command up docker container and run `main.go`.

```shell
make run
```

Now you can try [commands](#api-methods-description-) in your shell and see the results. Don't forget to create tables in database. File to create tables [here](./pkg/pgstore/create_table.sql)

To check tests 👇 This command up docker container, then run tests, and finally remove docker container.

```shell
make test
```

Other make command you can check [here](./Makefile). There is `lint`, `up`, `down` to manage project more satisfying.

## API methods description 📖

### addFunds (POST)

```shell
curl --location 'localhost:8080/api/v1/addFunds' \
--header 'Content-Type: application/json' \
--data '{
    "transactionID":"transaction-uuid-1",
    "userID":1,
    "balance":100
}'
```

#### Response

```json
{"id":3,"userID":1,"balance":100,"reserved":0,"updatedAt":"2023-03-28T17:52:16.152192+03:00"}
```

### reserveFunds (POST)

```shell
curl --location 'localhost:8080/api/v1/reserveFunds' \
--header 'Content-Type: application/json' \
--data '{
    "transactionID":"transaction-uuid-2",
    "walletID":1,
    "serviceID":1,
    "orderID":1,
    "price":15
}'
```

#### Response

```json
{"id":4,"walletID":1,"serviceID":1,"orderID":1,"price":15,"status":"","dateTime":"2023-03-28T17:57:41.681074+03:00"}
```

### recognizeRevenue (POST)

```shell
curl --location 'localhost:8080/api/v1/recognizeRevenue' \
--header 'Content-Type: application/json' \
--data '{
    "transactionID":"transaction-uuid-3",
    "walletID":1,
    "serviceID":1,
    "orderID":1,
    "status":"DONE"
}'
```

#### Response

```json
{"id":4,"walletID":1,"serviceID":1,"orderID":1,"price":15,"status":"DONE","dateTime":"2023-03-28T17:57:41.681074+03:00"}
```

### getUserBalance (POST)

```shell
curl --location --request GET 'localhost:8080/api/v1/getUserBalance' \
--header 'Content-Type: application/json' \
--data '{
    "userID":1
}'
```

#### Response

```json
{"id":3,"userID":1,"balance":100,"reserved":0,"updatedAt":"2023-03-28T17:52:16.152192+03:00"}
```