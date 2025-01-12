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
        scdoc
        go
      ];

      libPkgs = with pkgs; [
        # openssl_3
      ];

      devPkgs = with pkgs; [
        go
        just
        air
        sqlite
        templ
        tailwindcss
        glauth
        openldap
        hivemind
      ];
    in {
      packages.default = pkgs.buildGoModule {
        pname = "polybase";
        version = "0.1.0";
        src = ./.;
        # vendorHash = "sha256-mHW6g50nkVSuEOCKdis/N5qQxKrAsUxtCcooycqJRho=";

        nativeBuildInputs = buildPkgs;
        buildInputs = libPkgs;

        postInstall = ''
          mkdir -p $out/share/man/man1
          scdoc < polybase.1.scd > $out/share/man/man1/polybase.1
        '';
      };

      devShell = pkgs.mkShell {
        nativeBuildInputs = buildPkgs;
        buildInputs = libPkgs ++ devPkgs;
      };
    });
}
