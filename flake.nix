{
  description = "protoc-gen-template - A protoc plugin that renders arbitrary files from Go text/template sources driven by the parsed proto AST";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    { nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        version = (pkgs.lib.importJSON ./.github/config/release-please-manifest.json).".";
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "protoc-gen-template";
          inherit version;
          src = pkgs.lib.cleanSource ./.;
          subPackages = [ "cmd/protoc-gen-template" ];
          vendorHash = "sha256-0WAdrNgfeDAAoazHywCqjOkHB8QQ3KW+l2xvynG/zqk=";
          ldflags = [
            "-s"
            "-w"
          ];
          meta = with pkgs.lib; {
            description = "A protoc plugin that renders arbitrary files from Go text/template sources";
            license = licenses.mit;
            mainProgram = "protoc-gen-template";
          };
        };

        devShells.default = pkgs.mkShell {
          name = "protoc-gen-template";
          packages = [
            pkgs.go
            pkgs.protobuf
            pkgs.buf
          ];
        };
      }
    );
}
