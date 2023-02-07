# auto-restic

Automatic backup & restore. Extension of [Portable Stacks](https://reddec.net/articles/portable-stack/) ideas.

A tiny wrapper around [restic](https://github.com/restic/restic) with automatic initialization, restoring, and
backup.

Originally it was just bunch of shell scripts. But eventually, it logic became too complicated to fit bash.

Features

- Read-Only file system. No writing to file (except `/data`)
- Automatic **init & backup & restore**.
    - Container will be marked healthy only if repository initialized and data restored
- Separate cron-like tasks for pruning and backup with mutual lock
- Logs to stderr
- Supports readiness probe (healthcheck)
- Supports webhook notification

Example with S3:

```yaml
services:
  db:
    image: postgres:14
    environment:
      POSTGRES_PASSWORD: postgres
    volumes:
      - postgres:/var/lib/postgresql/data
    depends_on:
      backup:
        condition: service_healthy

  backup:
    build: ghcr.io/reddec/auto-restic:0.0.1
    environment:
      BACKUP_SCHEDULE: "@daily"
      RESTIC_PASSWORD: "backup-encryption-p@ssw0rd"
      RESTIC_REPOSITORY: "s3:https://s3.example.com/backups/${COMPOSE_PROJECT_NAME}"
      AWS_ACCESS_KEY_ID: "my-secret-key-id"
      AWS_SECRET_ACCESS_KEY: "my-secret-access-key"
      AWS_DEFAULT_REGION: us-west-000
    volumes:
      - postgres:/data/postgres

volumes:
  postgres: { }

```

> Note! It's designed to work as part of docker and never designed to be run as standalone application (however, it's
> still possible - just change all paths in configuration).

## Auto-build

Automatically builds & tests each restic releases.

Tests cover:

- auto initialization
- double initialization
- backup
- restore
- for local and S3-like repository


## Configuration

Configuration has reasonable defaults. Minimal required parameters:

- `RESTIC_REPOSITORY` - where to store backups
- `RESTIC_PASSWORD` - encryption password

Everything else is OPTIONAL. See [Dockerfile](./Dockerfile) for details.

## Shared volume

It uses `/data/.restored` marker file. Please keep user data (volumes) mounted under sub folders (usually by service
name) - see tests.

## Notification

Successful payload

```json
{
  "operation": "backup",
  "started": "2023-01-20T11:10:39.44006+08:00",
  "finished": "2023-01-20T11:10:39.751879+08:00",
  "failed": false
}
```

- operation could be: `backup` or `prune`

Failed payload (same as for success, but with `failed: true` and error message)

```json
{
  "operation": "backup",
  "started": "2023-01-20T11:10:39.44006+08:00",
  "finished": "2023-01-20T11:10:39.751879+08:00",
  "failed": true,
  "error": "something went wrong"
}
```
