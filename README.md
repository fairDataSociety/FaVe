## FaVe
[//]: # ([![Website]&#40;https://img.shields.io/badge/website-FAQ-orange?style=for-the-badge&#41;]&#40;https://fairdatasociety.github.io/FaVe/&#41;)
[//]: # (![Platform]&#40;https://img.shields.io/badge/platform-windows%20%7C%20macos%20%7C%20linux-green?style=for-the-badge&#41;)
[![Go Report Card](https://goreportcard.com/badge/github.com/fairDataSociety/FaVe?style=for-the-badge)](https://goreportcard.com/report/github.com/fairDataSociety/FaVe)
[![Release](https://img.shields.io/github/v/release/fairDataSociety/FaVe?include_prereleases&style=for-the-badge)](https://github.com/fairDataSociety/FaVe/releases)
![GitHub all releases](https://img.shields.io/github/downloads/fairDataSociety/FaVe/total?style=for-the-badge)
[![Workflow](https://img.shields.io/github/actions/workflow/status/fairDataSociety/FaVe/release.yaml?branch=master&style=for-the-badge)](https://github.com/fairDataSociety/FaVe/actions)
[![Issues](https://img.shields.io/github/issues-raw/fairDataSociety/FaVe?style=for-the-badge)](https://github.com/fairDataSociety/FaVe/issues)
[![Closed](https://img.shields.io/github/issues-closed-raw/fairDataSociety/FaVe?style=for-the-badge)](https://github.com/fairDataSociety/FaVe/issues?q=is%3Aissue+is%3Aclosed)
[![PRs](https://img.shields.io/github/issues-pr/fairDataSociety/FaVe?style=for-the-badge)](https://github.com/fairDataSociety/FaVe/pulls)
[![PRClosed](https://img.shields.io/github/issues-pr-closed-raw/fairDataSociety/FaVe?style=for-the-badge)](https://github.com/fairDataSociety/FaVe/pulls?q=is%3Apr+is%3Aclosed)
![Go](https://img.shields.io/github/go-mod/go-version/fairDataSociety/FaVe?style=for-the-badge&logo=go)
[![Discord](https://img.shields.io/discord/888359049551310869?style=for-the-badge&logo=discord)](https://discord.com/invite/KrVTmahcUA)
[![Telegram](https://img.shields.io/badge/-telegram-red?color=86d5f7&logo=telegram&style=for-the-badge)](https://t.me/joinchat/GCEfnpZbpfZgVyoK)
[![License](https://img.shields.io/badge/License-AGPL_v3-blue.svg?style=for-the-badge)](https://opensource.org/license/agpl-v3/)

FaVe is a truly decentralised, open source vector database build with Fair Data Principals in mind on top of FairOS. 

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
Collections can have documents with significant "properties" that can be "vectorized" and stored.

The properties of the documents are vectorized and stored in a specific document store in fairOS-dfs named after the 
collection itself. While adding documents in FaVe we calculate nearest neighbours. These are then stored in a key-value 
store in fairOS-dfs with unique key as identifiers.

Under the hood, everything is stored in SWARM via fairOS-dfs.

## What is Vectorizer?

Vectorizer is an http service that vectorizes the words in a given text. The openapi specs for a vectorizer can be found [here](./pkg/vectorizer/openapi-spec/schema.json).

Currently, there are two projects that can be used as vectorizer

- [glove-840B-leveldb](https://github.com/onepeerlabs/glove-840B-leveldb):  GloVe embeddings put inside leveldb.

- [huggingface-embeddings](https://github.com/fairDataSociety/huggingface-embeddings): Any Huggingface transformer can be loaded and served as FaVe vectorizer to generate text to embeddings.

## Running FaVe

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
go run cmd/fave-server/main.go --port 1234 --keep-alive 6000m --write-timeout 6000m --read-timeout 6000m
```

### With Docker

```
docker run -d \
    -e VERBOSE=true \
    -e BEE_API=<BEE_API> \
    -e RPC_API=<RPC_ENDPOINT_FOR_ENS_AUTH> \
    -e STAMP_ID=<STAMP_ID> \
    -e USER=<FAIROS_USERNAME> \
    -e PASSWORD=<FAIROS_PASSWORD> \
    -e POD=<POD_FOR_STORING_DB> \
    -e VECTORIZER_URL=<API_ENDPOINT_FOR_VECTORIZER> \
    -p 1234:1234 \
    fairdatasociety/fave:latest --port 1234 --host 0.0.0.0 --keep-alive 6000m --write-timeout 6000m --read-timeout 6000m
```

Or, you can build the docker image yourself.

```
// build 
docker build -t fds/fave .

// run
docker run -d \
    -e VERBOSE=true \
    -e BEE_API=<BEE_API> \
    -e RPC_API=<RPC_ENDPOINT_FOR_ENS_AUTH> \
    -e STAMP_ID=<STAMP_ID> \
    -e USER=<FAIROS_USERNAME> \
    -e PASSWORD=<FAIROS_PASSWORD> \
    -e POD=<POD_FOR_STORING_DB> \
    -e VECTORIZER_URL=<API_ENDPOINT_FOR_VECTORIZER> \
    -p 1234:1234 \
    fds/fave --port 1234 --host 0.0.0.0 --keep-alive 6000m --write-timeout 6000m --read-timeout 6000m

```

## Generate the server from openapi spec

```
go generate
```

## How does FaVe work?

FaVe currently supports only test vectorization.

The system first produces vector representations or embeddings from a chosen vectorizer. Following that, 
it determines the nearest neighbors based on these embeddings. Once this is done, the content gets uploaded, 
and subsequently, the information about the nearest neighbors is also uploaded.

When conducting a search for a particular term, it computes the distance from a designated starting point and 
then searches for a match within the precomputed nearest neighbors.

## How to put data in FaVe?

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

## How to search in FaVe?

FaVe provides a REST APIs for retrieving nearest documents from a collection, given a query and a maximum distance.

The response contains the nearest documents along with their properties and their distances from the query.

NOTE:
Please check this [issue](https://github.com/fairDataSociety/FaVe/issues/29) before running the server.

## Summary of the OpenAPI Specification
The `schema.json` file in the `openapi-spec` directory defines the OpenAPI specification for the FaVe project. Here's a summarized overview:

### General Information
- **Base Path**: `/v1`
- **Consumes**: `application/json`
- **Produces**: `application/json`
- **Swagger Version**: `2.0`
- **Contact**: `sabyasachi@datafund.io`
- **Project URL**: [GitHub Repository](https://github.com/fairDataSociety/FaVe)
- **Version**: `0.0.0-prealpha`

### Definitions
The schema defines several object types, including:
- `OKResponse`: A standard OK response.
- `ErrorResponse`: A standard error response.
- `Collection`: Represents a collection with a name and indexes.
- `Index`: Specifies fields to index in a collection.
- `Property`: An open object with additional properties.
- `Document`: Represents a document with properties and an ID.
- `AddDocumentsRequest`: Request object for adding documents to a collection.
- `NearestDocumentsRequest`: Request object for finding nearest documents by text.
- `NearestDocumentsByVectorRequest`: Request object for finding nearest documents by vector.
- `NearestDocumentsResponse`: Response object for nearest documents requests.

### API Endpoints
1. **Root Endpoint (`/`)**: A GET request returns a 200 status, indicating the API is alive.
  
2. **Collections (`/collections`)**:
    - GET: Retrieve all collections.
    - POST: Create a new collection.

3. **Specific Collection (`/collections/{collection}`)**:
    - DELETE: Delete a specific collection.

4. **Documents (`/documents`)**:
    - GET: Retrieve a document based on query parameters.
    - POST: Add documents to a collection.

5. **Nearest Documents (`/nearest-documents`)**:
    - POST: Get nearest documents for a collection based on text.

6. **Nearest Documents by Vector (`/nearest-documents-by-vector`)**:
    - POST: Get nearest documents for a collection based on a vector.

### Error Responses
The API has standard error responses like 400 (Bad Request), 401 (Unauthorized), 403 (Forbidden), 422 (Unprocessable Entity), and 500 (Internal Server Error).

### Tags
- `fave`: Everything about your fave.

