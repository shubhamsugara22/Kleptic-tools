# syntax=docker/dockerfile:1
# ─────────────────────────────────────────────────────────────────────────────
# Kleptic Tools — Custom Envoy image
#
# Build:
#   docker build -t kleptic/envoy:latest .
#   docker build --build-arg ENVOY_VERSION=v1.29-latest -t kleptic/envoy:v1.29 .
#
# Run (standalone):
#   docker run --rm -p 10000:10000 -p 9901:9901 kleptic/envoy:latest
#
# Run with custom config:
#   docker run --rm -p 10000:10000 -p 9901:9901 \
#     -v $(pwd)/my-envoy.yaml:/etc/envoy/envoy.yaml:ro \
#     kleptic/envoy:latest
# ─────────────────────────────────────────────────────────────────────────────

ARG ENVOY_VERSION=v1.29-latest
FROM envoyproxy/envoy:${ENVOY_VERSION}

# ── Labels ────────────────────────────────────────────────────────────────────
LABEL org.opencontainers.image.title="kleptic-envoy" \
	org.opencontainers.image.description="Custom Envoy proxy — Kleptic Tools" \
	org.opencontainers.image.source="https://github.com/kleptic/kleptic-tools" \
	org.opencontainers.image.licenses="Apache-2.0"

# ── System dependencies (minimal) ─────────────────────────────────────────────
# The official envoy image is Debian-based; install only what is needed.
USER root
RUN apt-get update && \
	apt-get install -y --no-install-recommends \
	ca-certificates \
	curl && \
	apt-get clean && \
	rm -rf /var/lib/apt/lists/*

# ── Config directory ──────────────────────────────────────────────────────────
RUN mkdir -p /etc/envoy && \
	chown -R envoy:envoy /etc/envoy

# Copy the default bootstrap configuration.
# Override at runtime by mounting your own config at /etc/envoy/envoy.yaml.
COPY envoy.yaml /etc/envoy/envoy.yaml

# Copy any additional files referenced from envoy.yaml (TLS certs placeholder
# paths, WASM filters, etc.).  These are no-ops if the files don't exist yet —
# remove or adjust as needed.
# COPY certs/           /etc/envoy/certs/
# COPY wasm/            /etc/envoy/wasm/

# ── Ports ─────────────────────────────────────────────────────────────────────
# 10000 — proxy listener (matches the default envoy.yaml)
# 9901  — admin API / Prometheus metrics
EXPOSE 10000 9901

# ── Switch to the non-root envoy user ─────────────────────────────────────────
USER envoy

# ── Health check ──────────────────────────────────────────────────────────────
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
	CMD curl -sf http://localhost:9901/ready || exit 1

# ── Entrypoint ────────────────────────────────────────────────────────────────
# --service-cluster and --service-node are consumed by tracing and xDS.
# Override SERVICE_CLUSTER / SERVICE_NODE via environment variables at runtime.
ENV SERVICE_CLUSTER=kleptic-envoy \
	SERVICE_NODE=envoy-1

ENTRYPOINT ["envoy", \
	"--config-path",     "/etc/envoy/envoy.yaml", \
	"--log-level",       "info", \
	"--service-cluster", "kleptic-envoy", \
	"--service-node",    "envoy-1"]

# To override the cluster/node at runtime:
#   docker run -e SERVICE_CLUSTER=my-cluster -e SERVICE_NODE=node-0 \
#     kleptic/envoy:latest \
#     --config-path /etc/envoy/envoy.yaml \
#     --service-cluster my-cluster \
#     --service-node node-0
