# Bitcoin Indexer with SQL Backend

The Bitcoin Indexer is a powerful tool that indexes blockchain data and provides efficient querying capabilities for exploring and analyzing Bitcoin transactions and addresses. This project utilizes a SQL backend for the database and leverages the Gorm library to connect with various backends. It exposes the same RPC methods as the original Bitcoin node, making it compatible with existing Bitcoin applications. 

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
- [Features](#features)
- [Architecture](#architecture)
- [Contributing](#contributing)
- [License](#license)

## Installation

1. Prerequisites: Go (version 1.18 or higher) should be installed.
2. Clone the repository: `git clone https://github.com/catalogfi/indexer.git`
3. Change to the project directory: `cd indexer`

## Usage

The Bitcoin Indexer provides an RPC server and a peer package for syncing blockchain data from other Bitcoin nodes.

1. **Peer Package**: The `cmd/peer` package can be used to connect to other Bitcoin nodes and sync the Bitcoin blockchain data. It handles the peer-to-peer communication required to fetch and update the blockchain data. 

   ```shell
   $ go build ./cmd/peer
   $ ./peer [db options]
   ```

2. **RPC Server**: The `cmd/rpc` package starts an RPC server that exposes the same RPC methods as the original Bitcoin node. This allows you to query the Bitcoin data using standard RPC calls.

   ```shell
   $ go build ./cmd/rpc
   $ ./rpc [db options]
   ```

   The RPC server can be scaled independently using a microservice-based architecture, allowing for cost-effective scalability when handling increased query loads.

## Features

- **Blockchain Indexing**: The Bitcoin Indexer efficiently indexes blockchain data using a SQL backend, providing fast and optimized querying capabilities.
- **Gorm Integration**: The project leverages the Gorm library, which supports various SQL backends, to connect with and interact with the database.
- **RPC Compatibility**: The indexer exposes the same RPC methods as the original Bitcoin node, making it compatible with existing Bitcoin applications.
- **Peer-to-Peer Sync**: The `cmd/peer` package allows you to sync blockchain data from other Bitcoin nodes, ensuring your indexer stays up to date with the latest Bitcoin transactions. 
- **Microservice Architecture**: The RPC server can be scaled independently, following a microservice-based architecture, allowing for cost-effective scalability as query loads increase.

## Architecture

The Bitcoin Indexer follows a modular architecture with the following folder structure:

- **cmd/peer**: This package handles peer-to-peer communication and syncing blockchain data from other Bitcoin nodes. It ensures your indexer stays synchronized with the Bitcoin network.

- **cmd/rpc**: This package starts an RPC server that exposes the same RPC methods as the original Bitcoin node. It provides a scalable solution for querying Bitcoin data and can be independently scaled as a microservice.

- **command**: This folder contains code to add new RPC methods to the indexer. It also includes the interface declaration for the storage object required by the RPC methods.

- **model**: The model folder defines the database structure compatible with GORM. It includes the necessary structs and mappings for interacting with the database.

- **peer**: The peer folder contains the code for connecting to other Bitcoin nodes, syncing data, retrieving newly discovered blocks, and submitting transactions. It handles the peer-to-peer communication required for blockchain synchronization.

- **rpc**: The rpc folder includes a basic GIN HTTP handler that listens for incoming RPC requests. It executes the commands defined in the command folder when specific RPC methods are triggered.

- **store**: The store folder implements the interfaces defined by the command and peer folders. It handles the actual storage and retrieval of data from the database.

```
indexer
├── cmd
│   ├── peer
│   │   ├── main.go
│   │   └── ...
│   └── rpc
│       ├── main.go
│       └── ...
├── command
│   ├── command.go
│   ├── codec.go
│   └── ...
├── model
│   ├── model.go
│   └── ...
├── peer
│   ├── peer.go
│   └── ...
├── rpc
│   ├── rpc.go
│   └── ...
└── store
│   ├── store.go
│   └── ...
├── config.yaml
├── README.md
└── ...
```

The modular architecture allows for separate development and testing of different components. It promotes code reusability and maintainability, enabling easy addition of new RPC methods, efficient syncing with other Bitcoin nodes, and scalable handling of RPC requests.

## Contributing

Contributions are welcome! If you would like to contribute to the Bitcoin Indexer project, please follow these steps:

1. Fork the repository.
2. Create a new branch: `git checkout -b feature/your-feature-name`
3. Make your changes and commit them: `git commit -am 'Add some feature'`
4. Push your changes to the branch: `git push origin feature/your-feature-name`
5. Submit a pull request.

Please ensure that your code adheres to the established coding style and includes appropriate tests for any new features or bug fixes.

## License

This project is licensed under the [GNU GPLv3](LICENSE). Feel free to use and modify the code according to the terms specified in the license.
