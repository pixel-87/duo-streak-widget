{ lib, dockerTools, cacert, default }:

dockerTools.buildLayeredImage {
  name = "duo-streak-widget";
  tag = "latest";

  contents = [
    cacert
    default
  ];

  config = {
    Entrypoint = [ "${default}/bin/duo-streak-widget" ];
    ExposedPorts = {
      "8080/tcp" = { };
    };
    Env = [
      "PORT=8080"
    ];
    WorkingDir = "/";
  };

  maxLayers = 120;
}
