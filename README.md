Full task [text](./docs/task.md), swagger.yaml [here](./api/swagger.yaml)

## API methods description 

### addFunds (POST)

```shell
curl --location 'localhost:8080/api/v1/addFunds' \
--header 'Content-Type: application/json' \
--data '{
    "transactionID":"transaction-uuid",
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
    "transactionID":"transaction-uuid",
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
    "transactionID":"transaction-uusid",
    "walletID":1,
    "serviceID":1,
    "orderID":1,
    "status":"DONE"
}'
```

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

```json
{"id":3,"userID":1,"balance":100,"reserved":0,"updatedAt":"2023-03-28T17:52:16.152192+03:00"}
```