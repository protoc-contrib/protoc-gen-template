{
  description = "protoc-gen-template - A protoc plugin that renders arbitrary files from Go text/template sources driven by the parsed proto AST";

  nixConfig = {
    extra-substituters = [ "https://nixos-contrib.cachix.org" ];
    extra-trusted-public-keys = [ "nixos-contrib.cachix.org-1:WRyXCyMQGjYEEC87RdlW4ZbK/R3Ia4/WGYEoIP9X/o0=" ];
  };

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      nixpkgs,
      flake-utils,
      ...
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        version = (pkgs.lib.importJSON ./.github/config/release-please-manifest.json).".";

        package = pkgs.buildGoModule {
          pname = "protoc-gen-template";
          inherit version;
          src = pkgs.lib.cleanSource ./.;
          subPackages = [ "cmd/protoc-gen-template" ];
          vendorHash = "sha256-hvCl5Ai6NSG/LECmof7WJt4805pVvjHlKQaYYHwgsQQ=";
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
        packages.default = package;

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
