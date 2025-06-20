# guardian

Use untrusted data stores with a dead simple secret manager CLI. Do not trust GDrive, S3 and really any non self-hosted service. Encrypt your files during writes

## Features

- [X] Encrypt secrets with a friendly CLI
- [X] Mount JSON secret database as a filesystem
- [ ] Share your secrets in a P2P, Zero Trust manner

## Install

```shell
go install github.com/RogueTeam/guardian/cmd/guardian@latest
```

## CLI

- CLI tool

```shell
guardian secrets help
```

Example:

```shell
guardian secrets [init get set list del]
```

- Mount (Linux only)

```shell
guardian mount help
```

Example:

```shell
guardian mount ./mountpoint
```

Then you could handle secret management as they where files in your system.
