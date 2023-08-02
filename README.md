## FaVeDB

FaVeDB is a truly decentralised, open source vector database build with Fair Data Principals in mind on top of FairOS. 

## Running FaVeDB

### From Source
Export the following from your terminal
```
export VERBOSE=                       
export BEE_API=
export RPC_API=
export STAMP_ID=
export LEVELDB_EMBEDDINGS_PATH=
export USER=
export PASSWORD=
export POD=
```

Note :
- VERBOSE is optional
- BEE_API is the url of the bee node
- RPC_API is the url of the ethereum (Sepolia for testnet) node for ENS authentication
- STAMP_ID is the id of the stamp you want to use uploading data into swarm
- LEVELDB_EMBEDDINGS_PATH is the local path for leveldb which has the embeddings as key-value pairs
- USER is the username of the user you want to use to access the database
- PASSWORD is the password of the user you want to use to access the database
- POD is the reference of the pod you want to use to store the database

Then run the following command
```
go run cmd/fave-server/main.go --port 1234
```

### With Docker

```
// build 
docker build -t fds/fave .

// run
docker run -d --name=fave \
    -v <LOCAL_LEVELDB_GLOVE_EMBEDDINGS_PATH>:/embeddings \
    -e VERBOSE=true \
    -e BEE_API=<BEE_API> \
    -e RPC_API=<RPC_ENDPOINT_FOR_ENS_AUTH> \
    -e STAMP_ID=<STAMP_ID> \
    -e USER=<FAIROS_USERNAMe> \
    -e PASSWORD=<FAIROS_PASSWORD> \
    -e POD=<POD_FOR_STORING_DB> \
    -e LEVELDB_EMBEDDINGS_PATH=/embeddings \
    -p 1234:1234 \
    fds/fave --port 1234 —host 0.0.0.0 

```

## Generate the server from openapi spec

```
go generate
```