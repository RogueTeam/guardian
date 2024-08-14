# guardian

Guardian is a JSON backed secret manager inspired in UNIX file experience.

## Features

- [X] Encrypt secrets with a friendly CLI
- [X] Mount JSON secret database as a filesystem
- [ ] Share your secrets in a P2P, Zero Trust manner

## CLI

- CLI tool

```shell
guardian secrets help
```shell

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
