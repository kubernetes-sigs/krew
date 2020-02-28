---
title: Installing
slug: install
weight: 200
---

Krew itself is a `kubectl` plugin that is installed and updated via Krew (yes,
Krew self-hosts).

> ⚠️ **Warning:** krew is only compatible with `kubectl` v1.12 or higher.

- **macOS/Linux**: [bash/zsh](#bash), [fish](#fish)
- **[Windows](#windows)**

## macOS/Linux {#posix}

#### Bash or ZSH shells {#bash}

1. Make sure that `git` is installed.
1. Run this command in your terminal to download and install `krew`:

    ```sh
    (
      set -x; cd "$(mktemp -d)" &&
      curl -fsSLO "https://github.com/kubernetes-sigs/krew/releases/latest/download/krew.{tar.gz,yaml}" &&
      tar zxvf krew.tar.gz &&
      KREW=./krew-"$(uname | tr '[:upper:]' '[:lower:]')_amd64" &&
      "$KREW" install --manifest=krew.yaml --archive=krew.tar.gz &&
      "$KREW" update
    )
    ```

1. Add `$HOME/.krew/bin` directory to your PATH environment variable. To do
   this, update your `.bashrc` or `.zshrc` file and append the following line:

     ```sh
     export PATH="${KREW_ROOT:-$HOME/.krew}/bin:$PATH"
     ```

   and restart your shell.

1. Verify running `kubectl krew` works.

#### Fish shell {#fish}

1. Make sure that `git` is installed.
1. Run this command in your terminal to download and install `krew`:

    ```fish
    begin
      set -x; set temp_dir (mktemp -d); cd "$temp_dir" &&
      curl -fsSLO "https://github.com/kubernetes-sigs/krew/releases/latest/download/krew.{tar.gz,yaml}" &&
      tar zxvf krew.tar.gz &&
      set KREWNAME krew-(uname | tr '[:upper:]' '[:lower:]')_amd64 &&
      ./$KREWNAME install \
        --manifest=krew.yaml --archive=krew.tar.gz &&
      set -e KREWNAME; set -e temp_dir
    end
    ```

1. Add `$HOME/.krew/bin` directory to your PATH environment variable. To do
   this, update your `config.fish` file and append the following line:

     ```fish
     set -gx PATH $PATH $HOME/.krew/bin
     ```

   and restart your shell.

1. Verify running `kubectl krew` works.

## Windows {#windows}

1. Make sure `git` is installed on your system.
1. Download `krew.exe` and `krew.yaml` from the [Releases][releases] page to
   a directory.
1. Launch a command-line window (`cmd.exe`) and navigate to that directory.
1. Run the following command to install krew (pass the correct
   paths to `krew.yaml` and `krew.zip` below):

    ```sh
    krew install --manifest=krew.yaml
    ```

1. Add `%USERPROFILE%\.krew\bin` directory to your `PATH` environment variable
   ([how?](https://java.com/en/download/help/path.xml))

1. Launch a new command-line window.
1. Verify running `kubectl krew` works.

[releases]: https://github.com/kubernetes-sigs/krew/releases

## Other package managers

You can alternatively install it via some OS-package managers like Homebrew
(macOS).

However, we don't actively support that scenario at the moment.
