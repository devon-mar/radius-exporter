#!/bin/bash
RC=0
set -eo pipefail
cd "$(dirname "$0")"

echo "Starting container"
echo "::group::Docker output"
docker run --rm -d --name freeradius -p 127.0.0.1:1812:1812/udp -v "$(pwd)/raddb/mods-config/files/authorize:/opt/etc/raddb/mods-config/files/authorize:ro" -v "$(pwd)/raddb/clients.conf:/opt/etc/raddb/clients.conf:ro" freeradius/freeradius-server:latest-alpine
sleep 5
echo "::endgroup::"

echo "Starting radius-exporter"
../radius-exporter &
PID=$!
RESP=$(curl --fail --silent --show-error "http://localhost:9881/metrics?target=127.0.0.1:1812&module=test")

echo "::group::Exporter Output"
echo "$RESP"
echo "::endgroup::"

echo "$RESP" | grep -q "^radius_success 1"
RET=$?
if [ $RET -ne 0 ]; then
    echo "::error::radius_success code wasn't 1."
    RC=$RET
fi

echo "$RESP" | grep -q "^radius_response_code 2"
RET=$?
if [ $RET -ne 0 ]; then
    echo "::error::Response code wasn't 2."
    RC=$RET
fi

echo "Killing radius-exporter"
kill $PID

echo "Killing free-radius"
docker rm -f freeradius

exit $RC
