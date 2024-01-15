FROM golang:1.20-alpine AS builder
WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN CGO_ENABLED=0 go build -v -o /usr/local/bin/auto-restic ./cmd/auto-restic

FROM restic/restic:0.16.3
COPY --from=builder /usr/local/bin/auto-restic /usr/bin/auto-restic
ENV \
    # Cron expression for backups
    BACKUP_SCHEDULE="@daily" \
    # Schedule to prune old backups
    BACKUP_PRUNE="@daily" \
    # How many old backups to keep during pruning
    BACKUP_DEPTH="7" \
    #
    # Notifications section
    #
    # Where to send notification, enabled only if URL is not empty
    BACKUP_NOTIFICATION_URL="" \
    # Number of bytes to store in memory for notification report
    BACKUP_LOG_LIMIT="8192" \
    # Number of retries to deliver notification
    BACKUP_NOTIFICATION_RETRIES="5" \
    # Interval between attempts
    BACKUP_NOTIFICATION_INTERVAL="12s" \
    # HTTP method for notifications
    BACKUP_NOTIFICATION_METHOD="POST" \
    # HTTP request timeout
    BACKUP_NOTIFICATION_TIMEOUT="30s" \
    # HTTP header (Authroization) for authroization
    BACKUP_NOTIFICATION_AUTHORIZATION="" \
    #
    # Internal configuration, most probably you should never edit it
    #
    # HTTP address used for internal API to detect that container is healthy
    API_ADDRESS="127.0.0.1:8080" \
    # Directory to backup/restore
    # Important! application will write .restored file there, so use subfolders to mount volumes (usually service name)
    BACKUP_DIR="/data"

# live data to backup/restore, should match BACKUP_DIR and MUST NOT be changed during restore
VOLUME [ "/data" ]
HEALTHCHECK --start-period=10s --retries=3 --interval=10s --timeout=3s CMD ["/usr/bin/auto-restic", "ready"]
ENTRYPOINT [ "" ]
CMD [ "/usr/bin/auto-restic", "run" ]