#!/bin/bash
# This script does an s3 cp for the data
csv="$(cat last_version_all.csv)"

mkdir -p temp

for line in ${csv}; do
    key="$(echo "${line}" | xsv 'select' 2)"
    id="$(echo "${line}" | xsv 'select' 3)"

    mkdir -p "temp/$(dirname "${key}")"

    aws s3api get-object --bucket frenchwhaling-subscribers --version-id "${id}" --key "${key}" "temp/${key}"
done

aws s3 sync temp/public s3://frenchwhaling-subscribers/private