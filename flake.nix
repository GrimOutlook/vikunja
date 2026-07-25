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

          version = self.shortRev or self.dirtyShortRev or "dev";

          # Plain relative paths are enough here - Nix filters a flake's
          # source to git-tracked content automatically when it's fetched
          # git-aware (accessed directly, or via another flake's `git+`
          # input). This does NOT hold for a `path:` input, which is a raw
          # directory copy with no .gitignore awareness; consume this flake
          # via git+file://<path> (or a real git remote), not path:.
          frontend = pkgs.stdenv.mkDerivation (finalAttrs: {
            pname = "vikunja-frontend";
            inherit version;
            src = ./frontend;

            # Some pnpm versions want to interactively confirm purging
            # node_modules before a clean install; there's no TTY in the
            # sandbox to answer that prompt.
            CI = "true";

            pnpmDeps = pkgs.fetchPnpmDeps {
              inherit (finalAttrs) pname version src;
              pnpm = pkgs.pnpm_10;
              fetcherVersion = 3;
              hash = "sha256-CA9clHM1X5hk7qMbCivNbe9jH36pR6pRZxF91BYD7fs=";
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
            src = ./.;

            vendorHash = "sha256-f9bFHZYqgGcrC13gZPb824hBQ4Ds2bf9wnCajZS+h3k=";

            nativeBuildInputs = [ pkgs.mage ];

            inherit frontend;
            prePatch = ''
              # Belt and suspenders against a stray pre-existing frontend/dist
              # (e.g. from a misconfigured source fetch): without this, `cp -r`
              # would nest ${frontend} inside it instead of replacing it,
              # breaking every go:embed'd asset path.
              rm -rf frontend/dist
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
