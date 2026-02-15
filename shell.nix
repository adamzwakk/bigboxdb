{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = [
    pkgs.go
    pkgs.nodejs_22
    
    pkgs.ktx-tools
    pkgs.vips

    pkgs.just
    pkgs.lazysql
    pkgs.zip
    pkgs.libwebp
  ];

  shellHook = ''
    echo "🐹 Go dev shell ready"
    go version
  '';
}
