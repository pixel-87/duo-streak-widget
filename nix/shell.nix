{
  mkShellNoCC,
  callPackage,

  # extra tooling
  go,
  gopls,
  goreleaser,
  terraform-ls,
  terraform,
}:
let
  defaultPackage = callPackage ./default.nix { };
in
mkShellNoCC {
  inputsFrom = [ defaultPackage ];

  packages = [
    go
    gopls
    goreleaser
    terraform-ls
    terraform
  ];
}
