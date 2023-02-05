# auto-restic

Automatic backup & restore.

Automatically builds & tests each restic release.

pros:

- RO-fs (except volume)
- notifications

test-case:

# init

- run & wait for boot
- insert data
- wait for 1 minute (to backup)
- stop

# double init

- run & wait for boot again
- check data
- stop

- remove volumes

# restore

- start & wait for boot
- check data

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
