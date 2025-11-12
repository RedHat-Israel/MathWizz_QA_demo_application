# MathWizz Message Queue Service
# NATS server for async event-driven communication

FROM docker.io/library/nats:2.10-alpine

# Copy NATS configuration
COPY nats.conf /etc/nats/nats.conf

# Create JetStream data directory
# Note: nats user already exists in the base image
RUN mkdir -p /data/jetstream

# Expose NATS ports
# 4222: Client connections
# 8222: HTTP monitoring
EXPOSE 4222 8222

# Health check
HEALTHCHECK --interval=10s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8222/healthz || exit 1

# Run NATS with our configuration
CMD ["nats-server", "-c", "/etc/nats/nats.conf"]
