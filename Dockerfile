# Build bot
FROM golang:1.21 AS builder
WORKDIR /src
ENV CGO_ENABLED=0
COPY . .
RUN make bot/build

# create user
FROM debian:bookworm-slim as debian
RUN useradd -u 1001 botuser && \
apt update && apt install -y ca-certificates

FROM scratch As base
# copy user
COPY --from=debian /etc/passwd /etc/passwd
# copy certs
COPY --from=debian /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=debian /etc/nsswitch.conf /etc/nsswitch.conf
ENV SSL_CERT_DIR=/etc/ssl/certs

# switch user
USER botuser
COPY --from=builder /src/zte-scanner-bot /zte-scanner-bot
ENV ZTE_BOT_TOKEN=""
ENV ZTE_BOT_ADMIN_ID=""
ENTRYPOINT ["/zte-scanner-bot","-db", "/db/devices.db", "-u", "admin", "-p", "admin"]
