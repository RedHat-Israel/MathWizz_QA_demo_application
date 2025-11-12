# MathWizz Database Service
# PostgreSQL container with schema initialization

FROM docker.io/library/postgres:16-alpine

# Set environment variables
ENV POSTGRES_DB=mathwizz
ENV POSTGRES_USER=mathwizz
ENV POSTGRES_PASSWORD=mathwizz_password

# Copy schema and initialization scripts
COPY schema.sql /docker-entrypoint-initdb.d/01-schema.sql
COPY init.sql /docker-entrypoint-initdb.d/02-init.sql

# Copy health check script
COPY healthcheck.sh /usr/local/bin/healthcheck.sh
RUN chmod +x /usr/local/bin/healthcheck.sh

# Configure health check
HEALTHCHECK --interval=10s --timeout=5s --start-period=30s --retries=3 \
    CMD /usr/local/bin/healthcheck.sh

# Expose PostgreSQL port
EXPOSE 5432

# The postgres image already has an ENTRYPOINT, so we don't need to specify one
