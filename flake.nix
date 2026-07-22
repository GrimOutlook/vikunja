{
  description = "Vikunja - a self-hosted to-do app (local fork build)";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-25.11";
  };

  outputs =
    { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
      ];
      forAllSystems = nixpkgs.lib.genAttrs systems;
    in
    {
      packages = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          inherit (pkgs) lib;

          # Track exactly what git tracks: excludes .gitignore'd build
          # artifacts (frontend/node_modules, the vikunja binary, *.db, etc.)
          # and .git itself, so unrelated local changes don't bust the cache.
          gitSource =
            root:
            lib.fileset.toSource {
              inherit root;
              fileset = lib.fileset.gitTracked root;
            };

          version = self.shortRev or self.dirtyShortRev or "dev";

          frontend = pkgs.stdenv.mkDerivation (finalAttrs: {
            pname = "vikunja-frontend";
            inherit version;
            src = gitSource ./frontend;

            pnpmDeps = pkgs.fetchPnpmDeps {
              inherit (finalAttrs) pname version src;
              pnpm = pkgs.pnpm_10;
              fetcherVersion = 3;
              hash = "sha256-3KEQTh6ye5Y5fyXmj43C5BH+IP+/PqwrI+uRvDVj5PA=";
            };

            nativeBuildInputs = [
              pkgs.nodejs_24
              pkgs.dart-sass
              pkgs.pnpmConfigHook
              pkgs.pnpm_10
            ];

            postBuild = ''
              # Force sass-embedded to use our dart-sass instead of bundled binaries.
              substituteInPlace node_modules/sass-embedded/dist/lib/src/compiler-path.js \
                --replace-fail 'compilerCommand = (() => {' 'compilerCommand = (() => { return ["${lib.getExe pkgs.dart-sass}"];'
              pnpm run build
            '';

            installPhase = ''
              cp -r dist/ $out
            '';
          });
        in
        {
          default = (pkgs.buildGoModule.override { go = pkgs.go_1_26; }) {
            pname = "vikunja";
            inherit version;
            src = gitSource ./.;

            vendorHash = "sha256-ukcWrTdc1yvP7i918INpqWLmDC7e/80IeUw0P5jbXqk=";

            nativeBuildInputs = [ pkgs.mage ];

            inherit frontend;
            prePatch = ''
              cp -r ${frontend} frontend/dist
            '';

            buildPhase = ''
              runHook preBuild

              # Fixes "mkdir /homeless-shelter: permission denied" during mage's own compile step.
              export HOME=$(mktemp -d)
              export RELEASE_VERSION=${version}
              mage build:build

              runHook postBuild
            '';

            doCheck = false;

            installPhase = ''
              runHook preInstall
              install -Dt $out/bin vikunja
              runHook postInstall
            '';

            meta = {
              description = "Todo-app to organize your life (local fork build)";
              homepage = "https://vikunja.io/";
              license = lib.licenses.agpl3Plus;
              mainProgram = "vikunja";
              platforms = lib.platforms.linux;
            };
          };
        }
      );
    };
}
