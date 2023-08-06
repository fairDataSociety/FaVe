## vector uploader

```go
go run uploader.go -b "http://localhost:1633" -s "51987f7304b419d8aa184d35d46b3cfeb1b00986ad937b3151c7ade699c81338" -d <POD_NAME> -u <USERNAME> -p <PASSWORD> -k <KV_STORE_NAME> -v <VECTOR_CSV> -x 1 -r <ENS_RPC_ENDPOINT>
```

### options available
```go
-l --verbose            Show fairos and other debug logs
-v --vector-csv-path    Path to the embedding file
-r --rpc-endpoint       RPC endpoint for ENS authentication
-b --bee-api-endpoint   Bee api endpoint
-s --stamp              stamp
-u --username           FDP portable username
-p --password           account password
-d --pod pod            name of the kv store
-k --kv-store           kv store name
-x --index-type         index type for the values in the kv store
```