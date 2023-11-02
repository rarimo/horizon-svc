# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
- Full config examples
- GET withdrawal by the deposit hash SSE endpoint
- `withdrawals` table to the database migrations
- Withdrawals cache querier

### Changed
- Refactored parsing of the genesis file for the collections and items indexers to omit race conditions and 
inconsistencies in the database
- Bridge Indexer to index the withdrawals events for the all chain types
- `txbuilder` package refactored to be able to create few transaction builders for the chains with the same type  

### Fixed
- Using UTC time in the all places
- Using one postgres connection for the all gorutines which could lock each other during execution database transactions
- `ChainTx` filter in the transfers select which could lead to the validation errors if the transaction hash is not in the
hex-encoded format
- `limit` filter in the transfers select which could lead to the empty response if the limit is not set

## [v1.0.0] - 2023-10-23
### Under the hood changes
- Initiated project

[Unreleased]: https://github.com/rarimo/horizon-svc/compare/v1.0.0...HEAD
[v1.0.0]: https://github.com/rarimo/horizon-svc/releases/tag/v1.0.0
