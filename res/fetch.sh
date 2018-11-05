#!/bin/bash

set -e

URL="https://downloads.asterisk.org/pub/telephony/sounds"
CORE="asterisk-core-sounds-en-wav-current"
EXTRA="asterisk-extra-sounds-en-wav-current"

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

