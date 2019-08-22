#!/bin/bash
mkdir -p merged

for f in $(find data/private -type f); do
	priv="${f}"
	pub="$(echo "${f}" | sed 's/private/public/g')"

	jq -n --argfile priv "${priv}" --argfile pub "${pub}" '{ first: $priv, last: $pub }' > "merged/$(basename $(dirname $priv)).json"
done
