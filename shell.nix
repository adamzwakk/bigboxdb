{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = [
    pkgs.go
    pkgs.ktx-tools

    pkgs.just
    pkgs.lazysql
  ];

  shellHook = ''
    echo "🐹 Go dev shell ready"
    go version
  '';
}
