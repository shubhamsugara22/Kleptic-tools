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

## Cheat Sheet

### Views

- `:po` pods
- `:svc` services
- `:deploy` deployments
- `:ns` namespaces
- `:node` nodes

### Common Actions

- `l` logs (on a pod)
- `d` describe (on a resource)
- `e` edit (on a resource)
- `x` shell/exec (on a pod)
- `s` scale (on a deployment)

### Filters and Search

- `/` filter list
- `Ctrl+f` find next match
- `Ctrl+b` find previous match

### Navigation

- `Esc` clear filter or exit modal
- `Enter` drill into a resource
- `Backspace` go back

## Notes

Make sure `kubectl` is installed and configured before using K9s.
