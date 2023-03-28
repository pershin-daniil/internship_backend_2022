You can read full task text [here](./docs/task.md), swagger.yaml [here](./api/swagger.yaml)

## API methods description 

### addFunds (POST)

```shell
curl --location 'localhost:8080/api/v1/addFunds' \
--header 'Content-Type: application/json' \
--data '{
    "transactionID":"transaction-uuid",
    "userID":1,
    "account_balance":100
}'
```

#### Response

```json
{"id":3,"userID":1,"account_balance":100,"updatedAt":"2023-03-27T12:07:33.352266+03:00"}
```