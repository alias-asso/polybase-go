#!/bin/sh
set -e

install -d -m 755 /etc/polybase
install -d -m 777 /var/log/polybase

install -m 644 dist/usr/local/man/man1/polybase.1 /usr/share/man/man1/
install -m 644 dist/usr/local/man/man1/polybased.1 /usr/share/man/man1/
makewhatis /usr/share/man

touch /etc/polybase/polybase.cfg

install -m 555 dist/etc/rc.d/polybased /etc/rc.d/

install -m 555 dist/usr/local/bin/polybase /usr/local/bin/
install -m 555 dist/usr/local/bin/polybased /usr/local/bin/

find migrations -name "*.sql" | sort | xargs cat | sqlite3 polybase.db

rcctl restart polybased
