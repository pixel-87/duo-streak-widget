{
  description = "Golang Project Template";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
  };

  outputs =
    { self, nixpkgs }:
    let
      forAllSystems =
        function:
        nixpkgs.lib.genAttrs nixpkgs.lib.systems.flakeExposed (
          system: function (import nixpkgs {
            inherit system;
            config.allowUnfree = true;
          })
        );
    in
    {
      packages = forAllSystems (pkgs: {
        widget = pkgs.callPackage ./nix/default.nix { };
        image = pkgs.callPackage ./nix/oci.nix {
          cacert = pkgs.cacert;
          default = self.packages.${pkgs.stdenv.hostPlatform.system}.widget;
        };
        default = self.packages.${pkgs.stdenv.hostPlatform.system}.widget;
      });

      devShells = forAllSystems (pkgs: {
        default = pkgs.callPackage ./nix/shell.nix { };
      });

      overlays.default = final: _: { widget = final.callPackage ./nix/default.nix { }; };
    };
}
