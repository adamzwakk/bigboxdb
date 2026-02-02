{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = [
    pkgs.go
    pkgs.nodejs_22
    pkgs.ktx-tools

    pkgs.just
    pkgs.lazysql
    pkgs.vips
  ];

  shellHook = ''
    echo "🐹 Go dev shell ready"
    go version
  '';
}
