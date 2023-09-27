# FaVe (`Fa`irOS `Ve`ctor store)

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

> **_IMPORTANT:_**  FaVe is under heavy development and in early BETA stage. Some abnormal behaviour, data loss can be observed. We do not recommend parallel usage of same account from multiple installations. Doing so might corrupt your data.

## How do I install FaVe?

## Prerequisites And Requirements

### Docker
You can get docker from [here](https://docs.docker.com/get-docker/)

### BEE
You will need a bee node running with a valid stamp id.

We encourage `Swarm Desktop` for setting up your bee node. Here is a [guide](https://medium.com/ethereum-swarm/upgrading-swarm-deskotp-app-beta-from-an-ultra-light-to-a-light-node-65d52cab7f2c) for it.

### FDP account
You will need a FDP/Fairdrive account to use FaVe. You can create one from [here](https://create-account.fairdatasociety.org/)

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
