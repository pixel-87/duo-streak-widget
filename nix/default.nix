{ lib, buildGoModule }:
let
  version = "0.0.1";
in
buildGoModule {
  pname = "duo-streak-widget";
  inherit version;

  src = ../src;

  vendorHash = "sha256-XWsBc8QKelO0FRwr2QgSNTNHyY9pD6fr1karRDc1CtY=";

  subPackages = [ "cmd/api" ];

  ldflags = [
    "-s"
    "-w"
    "-X main.version=${version}"
  ];

  postInstall = ''
    mv $out/bin/api $out/bin/duo-streak-widget
  '';

  meta = {
    description = "A duolingo 88x31 button";
    homepage = "https://github.com/pixel-87/duo-streak-widget";
    license = lib.licenses.mit;
    maintainers = with lib.maintainers; [ ];
    mainProgram = "duo-streak-widget";
  };
}
