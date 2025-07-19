#!/bin/sh
set -e

doas install -d -m 755 /etc/polybase
doas install -d -m 777 /var/log/polybase

doas install -m 644 dist/usr/local/man/man1/polybase.1 /usr/share/man/man1/
doas install -m 644 dist/usr/local/man/man1/polybased.1 /usr/share/man/man1/
doas makewhatis /usr/share/man

doas touch /etc/polybase/polybase.cfg

doas install -m 555 dist/etc/rc.d/polybased /etc/rc.d/

doas install -m 555 dist/usr/local/bin/polybase /usr/local/bin/
doas install -m 555 dist/usr/local/bin/polybased /usr/local/bin/

doas rcctl restart polybased
