#!/bin/sh

set -e

cp dist/usr/share/man/man1/polybase.1 dist/usr/share/man/man1/polybased.1 /usr/share/man/man1
makewhatis /usr/share/man

cp -r dist/etc/polybase /etc

cp dist/usr/local/bin/polybase dist/usr/local/bin/polybased /usr/local/bin
