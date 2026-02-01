{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = [
    pkgs.go
    pkgs.ktx-tools

    pkgs.just
    pkgs.lazysql
    pkgs.imagemagick
  ];

  shellHook = ''
    echo "🐹 Go dev shell ready"
    go version
  '';
}
