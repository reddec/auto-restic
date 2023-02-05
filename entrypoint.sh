#!/usr/bin/env bash

# which dir to backup
BACKUP_DIR="${BACKUP_DIR:-/data}"

# initialization error which indicates that repository already initialized
INIT_SUPPRESS=(
    "already initialized"
    "already exists"
)

# restore error which indicates that repository is empty
RESTORE_SUPPRESS=(
    "no snapshot found"
)

# check that the error belongs to already initialization error set
function is_double_init_error {
    for err in "${INIT_SUPPRESS[@]}"; do
        if [[ "$1" == *"$err"* ]]; then
            return 0
        fi
    done
    return 1
}

# check that the error belongs to no-snapshot-error set
function is_nothing_to_restore {
    for err in "${RESTORE_SUPPRESS[@]}"; do
        if [[ "$1" == *"$err"* ]]; then
            return 0
        fi
    done
    return 1
}

# initialize restic repository. can be called several times
if [ "$BACKUP_INIT" == "true" ]; then
    STDERR="$(restic init 2>&1 > /dev/null)"
    CODE=$?
    if [ $CODE != 0 ]; then
        if ! is_double_init_error "$STDERR"; then
            echo "$STDERR"
            echo "failed init"
            exit $CODE
        fi
    fi
fi

# restore data. it could be false-positve in case previous backup 
# failed, but in general - nothing bad if we will it run one more time.
if [ ! -f /restored ]; then
    echo "restoring..."
    # restore if possible
    STDERR=$(restic restore latest --target / 2>&1 > /dev/null)
    CODE=$?
    if [ $CODE != 0 ]; then
        if ! is_nothing_to_restore "$STDERR"; then
            echo "$STDERR"
            echo "failed restore"
            exit $CODE
        fi
        echo "nothing to restore"
    else
        echo "restored"
    fi
    touch /restored
fi

# we want to redirect logs via named pipes
PIPE=/run/log.pipe

rm -rf "$PIPE"
mkfifo $PIPE

echo "$BACKUP_SCHEDULE restic backup \"${BACKUP_DIR}\" 2>&1 >> $PIPE" > /etc/crontabs/root
crond -b -d 8

echo "ready"
exec tail -f "$PIPE"