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

## What is Vectorizer?

Vectorizer is an http service that vectorizes the words in a given text. The openapi specs for a vectorizer can be found [here](./pkg/vectorizer/openapi-spec/schema.json).

Currently, there are two projects that can be used as vectorizer

- [glove-840B-leveldb](https://github.com/onepeerlabs/glove-840B-leveldb):  GloVe embeddings put inside leveldb.

- [huggingface-embeddings](https://github.com/fairDataSociety/huggingface-embeddings): Any Huggingface transformer can be loaded and served as FaVe vectorizer to generate text to embeddings.

## Running FaVeDB

### From Source
Export the following from your terminal
```
export VERBOSE=                       
export BEE_API=
export RPC_API=
export STAMP_ID=
export VECTORIZER_URL=
export USER=
export PASSWORD=
export POD=
```

Note :
- VERBOSE is optional
- BEE_API is the url of the bee node
- RPC_API is the url of the ethereum (Sepolia for testnet) node for ENS authentication
- STAMP_ID is the id of the stamp you want to use uploading data into swarm
- VECTORIZER_URL is the vectorizer server url
- USER is the username of the user you want to use to access the database
- PASSWORD is the password of the user you want to use to access the database
- POD is the reference of the pod you want to use to store the database

`VECTORIZER_URL` is optional in case we want to provide embeddings generated from other sources

Then run the following command
```
go run cmd/fave-server/main.go --port 1234
```

### With Docker

```
docker run -d --name=fave \
    -e VERBOSE=true \
    -e BEE_API=<BEE_API> \
    -e RPC_API=<RPC_ENDPOINT_FOR_ENS_AUTH> \
    -e STAMP_ID=<STAMP_ID> \
    -e USER=<FAIROS_USERNAME> \
    -e PASSWORD=<FAIROS_PASSWORD> \
    -e POD=<POD_FOR_STORING_DB> \
    -e VECTORIZER_URL=<API_ENDPOINT_FOR_VECTORIZER> \
    -p 1234:1234 \
    fairdatasociety/fave:latest --port 1234 —host 0.0.0.0 
```

Or, you can build the docker image yourself.

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
    -e VECTORIZER_URL=<API_ENDPOINT_FOR_VECTORIZER> \
    -p 1234:1234 \
    fds/fave --port 1234 —host 0.0.0.0 

```

## Generate the server from openapi spec

```
go generate
```

## How does FaVeDB work?

FaVe currently supports only test vectorazation.

The system first produces vector representations or embeddings from a chosen vectorizer. Following that, 
it determines the nearest neighbors based on these embeddings. Once this is done, the content gets uploaded, 
and subsequently, the information about the nearest neighbors is also uploaded.

When conducting a search for a particular term, it computes the distance from a designated starting point and 
then searches for a match within the precomputed nearest neighbors.

## How to put data in FaVeDB?

FaVe utilizes fairOS internally, meaning it's embedded directly rather than through REST APIs. 
FaVe itself offers a set of REST APIs for various functions.

Before we go any further we need these concepts cleared up:

- Collection: This term refers to a namespace that points to a specific fairOS document and key-value store. 
  In user point of view a collection is a place to store documents.

- Documents: These are individual records placed within a collection.

- Properties: These represent features of a document that are stored. FaVe vectorizes a set of properties that can be used to for search.

- Vectorizer: It's worth noting that the vectorizer, responsible for creating vector representations of document properties, is intended to be a separate service from FaVe.

### How do we perform data upload?

FaVe provides a set of REST APIs for creating collections, adding documents, and retrieving nearest documents.

**Before uploading data as documents, some preprocessing is required.**

Here are the steps:

1. Start the vectorizer service.
2. Launch FaVe with fairOS credentials, the bee endpoint, and a batch.
3. Prepare the data for uploading.
4. Create a collection.
5. Upload the documents into FaVe.

We have to prepare the documents in a specific format before uploading them via REST api

```
{
  "name": "collection1",
  "propertiesToIndex": ["property1"],
  "documents": [
    {
      "id": "721dfcef-5b95-4eeb-99fc-841784a397df",
      "properties": {
        "property1": "foo1",
        "property2": "bar1"
      }
    },
    {
      "id": "721dfcef-5b95-4eeb-99fc-841784a397dg",
      "properties": {
        "property1": "foo2",
        "property2": "bar2"
      }
    }
  ]
}
```
This is an example of the add documents request body.

We have to provide the name of the collection. The propertiesToIndex is an array of properties that we want to index/vectorize in the vector database.
We are only indexing property1 in this example.

The documents array contains the documents that we want to upload. Each document has a unique id and properties. 
Properties are the features of the document. They should contain key and value pairs. all the documents should have the same properties.

Once we have the data in the correct format, we can upload it to FaVe.

## How to search in FaVeDB?

FaVe provides a REST APIs for retrieving nearest documents from a collection, given a query and a maximum distance.

The response contains the nearest documents along with their properties and their distances from the query.
