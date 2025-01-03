#!/bin/sh
set -e

install -d -m 755 /etc/polybase

install -m 644 dist/usr/local/man/man1/polybase.1 /usr/share/man/man1/
install -m 644 dist/usr/local/man/man1/polybased.1 /usr/share/man/man1/
makewhatis /usr/share/man

install -m 644 dist/etc/polybase/* /etc/polybase/

install -m 555 dist/etc/rc.d/polybased /etc/rc.d/

install -m 555 dist/usr/local/bin/polybase /usr/local/bin/
install -m 555 dist/usr/local/bin/polybased /usr/local/bin/
