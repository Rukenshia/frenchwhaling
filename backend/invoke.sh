#!/bin/bash
source .env.local

fn="${1}"

payload="$(cat functions/${fn}/resources/request.json | sed "s/{{ACCOUNT_ID}}/${ACCOUNT_ID}/" | sed "s/{{ACCESS_TOKEN}}/${ACCESS_TOKEN}/")"

serverless invoke local -f "${fn}" -e APPLICATION_ID="${APPLICATION_ID}" --data "${payload}"