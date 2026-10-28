{
  description = "Golang & React";

  inputs = {
    nixpkgs.url = "github:Nixos/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }:
    let
      system = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${system};
    in
    {
      devShells.${system}.default = pkgs.mkShell {
        buildInputs = with pkgs; [
          go
          gopls
          gotools
          golangci-lint

          nodejs_22
          typescript-language-server

          air
        ];

        shellHook = ''
          echo "Go Version     : $(go version)"
          echo "Node Version   : $(node -v)"
        '';
      };
    };
}    
