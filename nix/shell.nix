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
  google-cloud-sdk,
  gcc,
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
    # C compiler for cgo (some stdlib packages enable cgo on Linux)
    gcc
    terraform-ls
    terraform
    google-cloud-sdk
  ];
}
