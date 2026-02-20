# K9s

K9s is a terminal-based UI for managing Kubernetes clusters. It provides a fast, interactive way to view and operate on resources without leaving the CLI.

## Overview

K9s helps you:

- Browse Kubernetes resources quickly
- View logs and events in real time
- Edit manifests inline
- Switch contexts and namespaces easily
- Execute commands in pods

## Setup

### Install

Choose a method that matches your OS:

- **macOS (Homebrew)**
  ```bash
  brew install k9s
  ```

- **Windows (Chocolatey)**
  ```bash
  choco install k9s
  ```

- **Linux (Binary)**
  Download the latest release from:
  https://github.com/derailed/k9s/releases

### Verify

```bash
k9s version
```

### Run

```bash
k9s
```

K9s uses your current `kubectl` context and config from `~/.kube/config`.

## Useful Shortcuts

- `:` command mode
- `?` help and shortcuts
- `0` switch context
- `n` switch namespace
- `/` filter

## Notes

Make sure `kubectl` is installed and configured before using K9s.
