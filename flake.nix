{
    description = "protoc-go-gen-json";
    inputs = {
      nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    };
    outputs = inputs@{ nixpkgs, ... }:
    let
      system = "x86_64-linux";
      pkgs = import nixpkgs { inherit system; };
    in
    {
      devShells."${system}".default = import ./shell.nix { inherit pkgs; };
    };
}
