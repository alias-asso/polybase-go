{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = nixpkgs.legacyPackages.${system};

      buildPkgs = with pkgs; [
        pkg-config
        templ
        scdoc
        go
        tailwindcss
      ];

      devPkgs = with pkgs; [
        just
        air
        sqlite
        glauth
        openldap
        hivemind
      ];
    in {
      packages.default = pkgs.buildGoModule {
        pname = "polybase";
        version = "0.1.0";
        src = ./.;
        vendorHash = "sha256-WSBnlSVWfGZFQIyPk9ViA3uA6tbMqou8bcocAXYGul8=";

        nativeBuildInputs = buildPkgs;

        postPatch = ''
          tailwindcss -i static/css/main.css -o static/css/styles.css -m
          templ generate
        '';

        buildPhase = ''
          export GOOS=openbsd GOARCH=amd64 CGO_ENABLED=0
          mkdir -p bin
          go build -o bin/polybased ./polybased
          go build -o bin/polybase ./polybase
          scdoc < polybase.1.scd | sed "s/1980-01-01/$(date '+%B %Y')/" > polybase.1
          scdoc < polybased.1.scd | sed "s/1980-01-01/$(date '+%B %Y')/" > polybased.1
        '';

        installPhase = ''
          mkdir -p $out/dist/{usr/local/bin,usr/local/man/man1,etc/rc.d}
          cp bin/polybased bin/polybase $out/dist/usr/local/bin/
          cp *.1 $out/dist/usr/local/man/man1/
          cp polybased.rc $out/dist/etc/rc.d/polybased
          cp install.sh $out/
        '';
      };

      devShell = pkgs.mkShell {
        nativeBuildInputs = buildPkgs;
        buildInputs = devPkgs;
      };
    });
}
