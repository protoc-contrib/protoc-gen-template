{
  description = "protoc-gen-template - A protoc plugin that renders arbitrary files from Go text/template sources driven by the parsed proto AST";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    nix-release-bin = {
      url = "github:nixos-contrib/nix-release-bin";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.flake-utils.follows = "flake-utils";
    };
  };

  outputs =
    {
      nixpkgs,
      flake-utils,
      nix-release-bin,
      ...
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        version = (pkgs.lib.importJSON ./.github/config/release-please-manifest.json).".";

        source = pkgs.buildGoModule {
          pname = "protoc-gen-template";
          inherit version;
          src = pkgs.lib.cleanSource ./.;
          subPackages = [ "cmd/protoc-gen-template" ];
          vendorHash = "sha256-LXqGKEqveSG+esbGkSzjwF2Vsh7XGEG503WgTLyOzFo=";
          # The standard library only uses cgo for net and os/user, which fall
          # back to pure Go without it. The Linux binary is then static, so the
          # release asset runs on any distribution, not only under Nix.
          env.CGO_ENABLED = 0;
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
      in
      {
        packages = {
          # The latest release binary, where it has one for the system: CI pins
          # them in the manifest once the release has published them.
          default = nix-release-bin.lib.mkReleaseBin {
            inherit pkgs;
            manifest = ./.github/config/nix-release-bin-manifest.json;
            pname = "protoc-gen-template";
            fallback = source;
          };
          inherit source;
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
