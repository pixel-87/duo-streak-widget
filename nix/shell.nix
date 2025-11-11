{
  mkShellNoCC,
  callPackage,

  # extra tooling
  go,
  gopls,
  golangci-lint,
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
    golangci-lint
    goreleaser
    terraform-ls
    terraform
  ];
}
