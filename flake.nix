{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-24.11";
    flake-utils.url = "github:numtide/flake-utils";
  };
  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = import nixpkgs {
        inherit system;
      };
      fastlauncher-package = pkgs.callPackage ./package.nix {};
    in {
      packages = rec {
        fastlauncher = fastlauncher-package;
        default = fastlauncher;
      };

      apps = rec {
        fastlauncher = flake-utils.lib.mkApp {
          drv = self.packages.${system}.fastlauncher;
        };
        default = fastlauncher;
      };

      devShells.default = pkgs.mkShell {
        packages = (with pkgs; [
          go
        ]);
      };
    });
}
