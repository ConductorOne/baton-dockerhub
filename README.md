![Baton Logo](./docs/images/baton-logo.png)

# `baton-dockerhub` [![Go Reference](https://pkg.go.dev/badge/github.com/conductorone/baton-dockerhub.svg)](https://pkg.go.dev/github.com/conductorone/baton-dockerhub) ![main ci](https://github.com/conductorone/baton-dockerhub/actions/workflows/main.yaml/badge.svg)

`baton-dockerhub` is a connector for DockerHub built using the [Baton SDK](https://github.com/conductorone/baton-sdk). It communicates with the DockerHub API to sync data about organizations, teams, users and repositories. 

To make connection with the DockerHub user provisioning API, we used mechanism for obtaining credentials from docker CLI tool [hub-tool](https://github.com/docker/hub-tool), you can find used code and changes done on it in [folder external](./pkg/dockerhub/external/).

Check out [Baton](https://github.com/conductorone/baton) to learn more about the project in general.

# Prerequisites

Among the prerequisities for running `baton-dockerhub` are prerequisities for running [hub-tool](https://github.com/docker/hub-tool#prerequisites) which is installed Docker on your machine and DockerHub account. You can use account username and access token to authenticate in connector.

# Getting Started

## brew

```
brew install conductorone/baton/baton conductorone/baton/baton-dockerhub

BATON_USERNAME=username BATON_ACCESS_TOKEN=access-token baton-dockerhub
baton resources
```

## docker

```
docker run --rm -v $(pwd):/out -e BATON_USERNAME=username BATON_ACCESS_TOKEN=access-token ghcr.io/conductorone/baton-dockerhub:latest -f "/out/sync.c1z"
docker run --rm -v $(pwd):/out ghcr.io/conductorone/baton:latest -f "/out/sync.c1z" resources
```

## source

```
go install github.com/conductorone/baton/cmd/baton@main
go install github.com/conductorone/baton-dockerhub/cmd/baton-dockerhub@main

BATON_USERNAME=username BATON_ACCESS_TOKEN=access-token baton-dockerhub
baton resources
```

# Data Model

`baton-dockerhub` will pull down information about the following DockerHub resources:

- Organizations
- Teams
- Users
- Repositories

By default, `baton-dockerhub` will sync information from all available organizations, but you can also specify exactly which organizations you would like to sync using the `--orgs` flag.

# Contributing, Support and Issues

We started Baton because we were tired of taking screenshots and manually building spreadsheets. We welcome contributions, and ideas, no matter how small -- our goal is to make identity and permissions sprawl less painful for everyone. If you have questions, problems, or ideas: Please open a Github Issue!

See [CONTRIBUTING.md](https://github.com/ConductorOne/baton/blob/main/CONTRIBUTING.md) for more details.

# `baton-dockerhub` Command Line Usage

```
baton-dockerhub

Usage:
  baton-dockerhub [flags]
  baton-dockerhub [command]

Available Commands:
  capabilities       Get connector capabilities
  completion         Generate the autocompletion script for the specified shell
  help               Help about any command

Flags:
      --access-token string    The DockerHub Personal Access Token used to connect to the DockerHub API. ($BATON_ACCESS_TOKEN)
      --client-id string       The client ID used to authenticate with ConductorOne ($BATON_CLIENT_ID)
      --client-secret string   The client secret used to authenticate with ConductorOne ($BATON_CLIENT_SECRET)
  -f, --file string            The path to the c1z file to sync with ($BATON_FILE) (default "sync.c1z")
  -h, --help                   help for baton-dockerhub
      --log-format string      The output format for logs: json, console ($BATON_LOG_FORMAT) (default "json")
      --log-level string       The log level: debug, info, warn, error ($BATON_LOG_LEVEL) (default "info")
      --orgs strings           Limit syncing to specific organizations by providing organization slugs. ($BATON_ORGS)
      --password string        The DockerHub password used to connect to the DockerHub API. ($BATON_PASSWORD)
  -p, --provisioning           This must be set in order for provisioning actions to be enabled ($BATON_PROVISIONING)
      --skip-full-sync         This must be set to skip a full sync ($BATON_SKIP_FULL_SYNC)
      --ticketing              This must be set to enable ticketing support ($BATON_TICKETING)
      --username string        required: The DockerHub username used to connect to the DockerHub API. ($BATON_USERNAME)
  -v, --version                version for baton-dockerhub

Use "baton-dockerhub [command] --help" for more information about a command.
```
