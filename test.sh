#!/usr/bin/env bash
set -e -x -o pipefail

INTERVAL="1"
CASE="${1:-local}"
export COMPOSE_FILE="tests/$CASE/docker-compose.yaml"

function wait_for_start {
  # count expected
  expected_count="$(docker compose ps -qa | wc -l | tr -d ' ')"
  # wait for boot
  echo "wait for boot..."
  for i in {1..120}; do
    echo "attempt $i - checking containers"
    if [ "$(docker compose ps -q | wc -l | tr -d ' ')" == "${expected_count}" ]; then
      return
    fi
    sleep "$INTERVAL"
  done
  exit 1
}

function wait_for_postgres {
  for i in {1..120}; do
    if [ "$(docker compose exec db ls -A /var/run/postgresql)" ]; then
      return
    fi
    sleep "$INTERVAL"
  done
  exit 1
}

function validate_data {
  # test data result
  EXPECTED='id,value
1,foo
2,bar
(2 rows)'

  DATA="$(echo 'SELECT * FROM test_data ORDER BY id' | docker compose exec -T -e PGPASSWORD=postgres db psql -U postgres -d postgres -A -F',')"
  if [ "$EXPECTED" != "$DATA" ]; then
    echo "data was corrupted"
    exit 2
  fi
}

# clean up

if [ "$(docker ps -q)" ]; then
  docker ps -q | xargs -n 1 docker stop
fi

if [ "$(docker ps -qa)" ]; then
  docker ps -qa | xargs -n 1 docker rm
fi

if [ "$(docker volume ls -q)" ]; then
  docker volume ls -q | xargs -n 1 docker volume rm
fi
rm -rf tests/$CASE/.data

# first init
docker compose build
docker compose create
docker compose up -d
wait_for_start
wait_for_postgres

# add data to DB

docker compose exec -T -e PGPASSWORD=postgres db psql -U postgres -d postgres <<EOF
CREATE TABLE test_data (
    id BIGSERIAL NOT NULL PRIMARY KEY,
    value TEXT NOT NULL
);

INSERT INTO test_data(value) VALUES ('foo'), ('bar');
EOF

# wait for backup
echo "waiting for backup..."
sleep 60

# stop
docker compose stop

# case - double init
echo "testing double initialization"
docker compose create
docker compose up -d
wait_for_start
wait_for_postgres
validate_data

# case - restore

# cleanup all except dir with snapshots
docker ps -q | xargs -n 1 docker stop
docker ps -qa | xargs -n 1 docker rm
docker volume ls -q | xargs -n 1 docker volume rm

echo "testing recovery"
docker compose create
docker compose up -d
wait_for_start
wait_for_postgres
validate_data
