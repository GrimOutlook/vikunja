{ pkgs, lib, config, inputs, ... }:

let
  pkgs-unstable = import inputs.nixpkgs-unstable { system = pkgs.stdenv.system; };
in {
  scripts.patch-sass-embedded.exec = ''
  find node_modules/.pnpm/sass-embedded-linux-*/node_modules/sass-embedded-linux-*/dart-sass/src -name dart -print0 | xargs -I {} -0 patchelf --set-interpreter "$(<$NIX_CC/nix-support/dynamic-linker)" {}
  '';

	packages = with pkgs-unstable; [
    # General tools
    git-cliff 
    actionlint
    crowdin-cli
    nfpm
    # API tools
    golangci-lint mage
    # Desktop
    electron
    # Font processing tools
    wget
    python3
    python3Packages.pip
    python3Packages.fonttools
    python3Packages.brotli
    nodejs
  ];
  
  languages = {
    javascript = {
      enable = true;
      package = pkgs-unstable.nodejs-slim;
      pnpm = {
        enable = true;
        package = pkgs-unstable.pnpm;
      };
    };
    
    go = {
      enable = true;
      package = pkgs-unstable.go;
      enableHardeningWorkaround = true;
    };
  };
  
  services.mailpit = {
    enable = true;
    package = pkgs-unstable.mailpit;
  };

  # `devenv up` starts both in one call. `mage build:build` stubs
  # frontend/dist first so the API's //go:embed doesn't fail on a fresh
  # checkout; the frontend installs/patches sass-embedded before serving.
  processes = {
    api.exec = "mage build:build && ./vikunja";
    frontend.exec = "cd frontend && pnpm install && patch-sass-embedded && pnpm dev";
  };

  env = {
    PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD = "1";
    PLAYWRIGHT_SKIP_VALIDATE_HOST_REQUIREMENTS = "1";
#    PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH = "${pkgs-unstable.chromium}/bin/chromium";
    VIKUNJA_SERVICE_TESTINGTOKEN = "test";
    # Required by the api process below: CORS is on by default and refuses to
    # start without a public URL configured, and a fresh checkout has no config.yml.
    VIKUNJA_SERVICE_PUBLICURL = "http://localhost:3456/";
  };
	
	devcontainer = {
		enable = true;
		settings = {
			forwardPorts = [ 4173 3456 ];
			portsAttributes = {
				"4173" = {
					label = "Vikunja Frontend dev server";
				};
				"3456" = {
					label = "Vikunja API";
				};
			};
			customizations.vscode.extensions = [
        "Syler.sass-indented"
        "codezombiech.gitignore"
        "dbaeumer.vscode-eslint"
        "editorconfig.editorconfig"
        "golang.Go"
        "lokalise.i18n-ally"
        "mikestead.dotenv"
        "mkhl.direnv"
        "vitest.explorer"
        "vue.volar"
			];
		};
	};
}
