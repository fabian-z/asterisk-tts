#!/bin/bash

set -e

format="$1"
if [[ "$format" != "wav" && "$format" != "sln16" ]]; then
	echo "Unsupported format: $1"
	exit
fi

URL="https://downloads.asterisk.org/pub/telephony/sounds"
CORE="asterisk-core-sounds-en-$1-current"
EXTRA="asterisk-extra-sounds-en-$1-current"

rm -rf "$CORE" && rm -rf "$EXTRA" 
mkdir "$CORE" && mkdir "$EXTRA"

wget "$URL/$CORE.tar.gz"
wget "$URL/$EXTRA.tar.gz"

cd "$CORE"
tar xvf "../$CORE.tar.gz"

cd ..
cd "$EXTRA"
tar xvf "../$EXTRA.tar.gz"

cd ..
cp "$CORE/core-sounds-en.txt" ./
cp "$EXTRA/extra-sounds-en.txt" ./

