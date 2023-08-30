## FaVeDB

FaVeDB is a truly decentralised, open source vector database build with Fair Data Principals in mind on top of FairOS. 

## Architecture

```
               FaVe
┌───────────────────────────────────┐
│                                   │
│ ┌────┐ ┌────┐ ┌────┐       ┌────┐ │
│ │ c1 │ │ c2 │ │ c3 │ x x x │ cn │ │               ┌──────────────────────┐
│ └─┬──┘ └─┬──┘ └──┬─┘       └─┬──┘ ├──────────────►│                      │
│   │      │       │           │    │               │      Vectorizer      │
│   │      │       │           │    │◄──────────────┤                      │
│   │      │       │           │    │               └──────────────────────┘
│   │      │       ▼           │    │
│   ▼      ▼   FairOS-dfs      ▼    │
│  ┌────┬────────────────────────┐  │
│  │    │                        │  │
│  │ ┌──▼────────┐ ┌───────────┐ │  │
│  │ │  document │ │ key-value │ │  │
│  │ │   store   ├─►   store   │ │  │
│  │ └─────┬─────┘ └─────┬─────┘ │  │
│  │       │             │       │  │
│  │  ┌────▼─────────────▼────┐  │  │
│  │  │         SWARM         │  │  │
│  │  └───────────────────────┘  │  │
│  │                             │  │
│  └─────────────────────────────┘  │
│                                   │
└───────────────────────────────────┘
```

In the diagram we can see c1, c2, c3, which are collections. We can have multiple collections in a single FaVe instance.
Collections can have documents with significant "Properties" that can be "vectorized" and stored.

The properties of the documents are vectorized and stored in a specific document store in fairOS-dfs named after the 
collection itself. While adding documents in FaVe we calculate nearest neighbours. These are then stored in a key-value 
store in fairOS-dfs with unique key as identifiers.

Under the hood, everything is stored in SWARM via fairOS-dfs.

## Running FaVeDB

### From Source
Export the following from your terminal
```
export VERBOSE=                       
export BEE_API=
export RPC_API=
export STAMP_ID=
export GLOVE_LEVELDB_URL=
export USER=
export PASSWORD=
export POD=
```

Note :
- VERBOSE is optional
- BEE_API is the url of the bee node
- RPC_API is the url of the ethereum (Sepolia for testnet) node for ENS authentication
- STAMP_ID is the id of the stamp you want to use uploading data into swarm
- GLOVE_LEVELDB_URL is glove vectorizer url
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
    -e VERBOSE=true \
    -e BEE_API=<BEE_API> \
    -e RPC_API=<RPC_ENDPOINT_FOR_ENS_AUTH> \
    -e STAMP_ID=<STAMP_ID> \
    -e USER=<FAIROS_USERNAME> \
    -e PASSWORD=<FAIROS_PASSWORD> \
    -e POD=<POD_FOR_STORING_DB> \
    -e GLOVE_LEVELDB_URL=<API_ENDPOINT_FOR_VECTORIZER> \
    -p 1234:1234 \
    fds/fave --port 1234 —host 0.0.0.0 

```

## Generate the server from openapi spec

```
go generate
```

## What is Vectorizer?

Vectorizer is a service that vectorizes the words in a given text. 

Currently, the only supported vectorizer is GloVe put inside leveldb (https://github.com/onepeerlabs/glove-840B-leveldb). 
It uses GloVe word embeddings to vectorize the words.

We can use any other pretrained word embeddings to vectorize the words. We just need to put the embeddings in a leveldb 
and use this script https://github.com/onepeerlabs/glove-840B-leveldb/tree/master/cmd/csvToLeveldb to convert the embeddings to leveldb.