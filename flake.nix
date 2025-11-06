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
        example = pkgs.callPackage ./nix/default.nix { };
        default = self.packages.${pkgs.stdenv.hostPlatform.system}.example;
      });

      devShells = forAllSystems (pkgs: {
        default = pkgs.callPackage ./nix/shell.nix { };
      });

      overlays.default = final: _: { example = final.callPackage ./nix/default.nix { }; };
    };
}
