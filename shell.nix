{ pkgs ? import <nixpkgs> { } }:

let
  protobuf-go = pkgs.buildGoModule rec {
    pname = "protobuf-go";
    version = "v1.28.1";

    src = pkgs.fetchFromGitHub {
      owner = "protocolbuffers";
      repo = "protobuf-go";
      rev = version;
      sha256 = "sha256-7Cg7fByLR9jX3OSCqJfLw5PAHDQi/gopkjtkbobnyWM=";
    };

    vendorSha256 = "sha256-yb8l4ooZwqfvenlxDRg95rqiL+hmsn0weS/dPv/oD2Y=";
    subPackages = [ "cmd/protoc-gen-go" ];
  };
in pkgs.mkShell rec {
  name = "protoc-gen-go-json";
  buildInputs = [
    pkgs.go_1_18
    pkgs.protobuf3_21
    protobuf-go
  ];
}
