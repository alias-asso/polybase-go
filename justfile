dev:
  hivemind

dev-rw:
  sudo sh jammer.sh start
  hivemind; sudo sh jammer.sh stop

build: clean build-server build-cli

publish: test build
  mkdir -p target/dist/{usr/local/bin,usr/share/man/man1,etc/polybase,etc/rc.d}
  cp target/polybased target/dist/usr/local/bin
  cp target/polybase target/dist/usr/local/bin
  cp target/polybase.1 target/dist/usr/share/man/man1/
  cp target/polybased.1 target/dist/usr/share/man/man1/
  touch target/dist/etc/polybase/polybase.cfg
  cp polybased.rc target/dist/etc/rc.d/polybased
  cp install.sh target/dist/
  cd target && tar czf dist.tar.gz dist

test:
  go test -cover ./...

migrate:
  find migrations -name "*.sql" | sort -n | xargs cat | sqlite3 polybase.db

clean:
  rm -fr .cache/
  rm -fr target/

build-server:
  mkdir -p target
  tailwindcss -i static/css/main.css -o static/css/styles.css -m
  templ generate
  go build -o target/polybased ./polybased
  scdoc < polybased.1.scd | sed "s/1980-01-01/$(date '+%B %Y')/" > target/polybased.1

build-cli:
  mkdir -p target
  go build -o target/polybase ./polybase
  scdoc < polybase.1.scd | sed "s/1980-01-01/$(date '+%B %Y')/" > target/polybase.1

build-docker:
  nix build .#docker
  docker load < result
