{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = [
    pkgs.go
  ];

  shellHook = ''
    echo "🐹 Go dev shell ready"
    go version
  '';
}
