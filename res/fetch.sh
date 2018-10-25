#!/bin/bash
wget https://downloads.asterisk.org/pub/telephony/sounds/asterisk-core-sounds-en-wav-current.tar.gz
wget https://downloads.asterisk.org/pub/telephony/sounds/asterisk-extra-sounds-en-wav-current.tar.gz
tar xvf asterisk-core-sounds-en-wav-current.tar.gz core-sounds-en.txt
tar xvf asterisk-extra-sounds-en-wav-current.tar.gz extra-sounds-en.txt
