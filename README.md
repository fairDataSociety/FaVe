## FaVeDB

FaVeDB is an open source vector database build with Fair Data Principals in mind on top of FairOS. 

## Running FaVeDB

Export the following from your terminal
```
export VERBOSE=                       
export BEE_API=
export RPC_API=
export STAMP_ID=
export GLOVE_POD_REF=
export USER=
export PASSWORD=
export POD=
```

Note :
- VERBOSE is optional
- BEE_API is the url of the bee node
- RPC_API is the url of the ethereum (Sepolia for testnet) node for ENS authentication
- STAMP_ID is the id of the stamp you want to use uploading data into swarm
- GLOVE_POD_REF is the reference of the pod containing the glove model
- USER is the username of the user you want to use to access the database
- PASSWORD is the password of the user you want to use to access the database
- POD is the reference of the pod you want to use to store the database

Then run the following command
```
go run cmd/fave-server/main.go --port 1234
```

## Using FaVeDB