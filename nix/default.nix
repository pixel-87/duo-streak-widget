{ lib, buildGoModule }:
let
  version = "0.0.1";
in
buildGoModule {
  pname = "duo-streak-widget";
  inherit version;

  src = ./src/;

  vendorHash = null;

  ldflags = [
    "-s"
    "-w"
    "-X main.version=${version}"
  ];

  meta = {
    description = "A duolingo 88x31 button";
    homepage = "https://github.com/pixel-87/duo-streak-widget";
    license = lib.licenses.mit;
    maintainers = with lib.maintainers; [ ];
    mainProgram = "example";
  };
}
