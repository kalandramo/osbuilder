# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`osbuilder` is a CLI scaffolding tool for the [onexstack](https://github.com/onexstack/onexstack) technology stack. It generates production-ready Go projects and can add REST API resources, CLI commands, async jobs, or message queue consumers to existing projects.

## Common Commands

```bash
make all              # tidy + format + build (default)
make build            # build binary for current platform → _output/platforms/{OS}/{ARCH}/
make build.multiarch  # build for all supported platforms
make format           # format with gofmt
make tidy             # go mod tidy
make clean            # remove _output/
go test ./...         # run all tests
go test -v -run TestName ./path/to/pkg  # run a single test
```

## Architecture

### Entry Point

`cmd/osbuilder/osbuilder.go` → `internal/osbuilder/cmd/cmd.go` (`NewDefaultOSCtlCommand`). Uses Kubernetes CLI patterns (cobra + k8s.io/cli-runtime).

### Core Packages

| Package | Role |
|---|---|
| `internal/osbuilder/cmd/` | All Cobra command implementations |
| `internal/osbuilder/cmd/create/` | `create project`, `create api`, `create cmd`, `create job`, `create mq`, `create quickstart` |
| `internal/osbuilder/cmd/semver/` | Semantic versioning and release management (`tag`, `release`) |
| `internal/osbuilder/types/` | Config structs for project generation (Project, WebServer, JobServer, MQServer, CLITool) |
| `internal/osbuilder/tpl/` | Embedded Go templates for generated projects (`project/`, `api/`, `cmd/`) |
| `internal/osbuilder/semver/` | Semver engine: version parsing, uplift, changelog, GPG signing, task execution |
| `internal/osbuilder/util/` | Shared helpers: file ops, template rendering, signal handling |
| `internal/osbuilder/statik/` | Embedded static assets |
| `internal/osbuilder/validation/` | Input validation for project config |

### Project Generation Flow

1. User provides `project.yaml` (or falls back to defaults in `tpl/project.yaml`)
2. `create project` validates the config via `internal/osbuilder/types/` and `validation/`
3. Templates under `internal/osbuilder/tpl/project/` are rendered with project metadata
4. The generated project includes: `cmd/`, `internal/`, `pkg/`, Makefile, Dockerfile, YAML configs, and example code matching the chosen application types
5. Subsequent `create api/cmd/job/mq` commands append new resources to an existing generated project

### Key Design Patterns

- **Template rendering**: Templates live in `tpl/` and are embedded via `statik`. Helper logic is in `internal/osbuilder/util/templates/`.
- **Plugin system**: `cmd.go` supports kubectl-style plugin discovery from PATH.
- **Semver tasks**: Release pipeline is composed of discrete tasks in `semver/task/` (bump, tag, changelog, GPG sign, GitHub publish).
