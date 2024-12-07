# hxutil

This is a utility I've made for automating and simplifying workflows and testing with Hexabase.

It has a few useful tools so far, mainly for managing ActionScript and testing APIs. In the future, hopefully it'll include a lot more.

## Installation

First, you need to have Go installed. See the Go website for more info.

If you just want to use the CLI, the easiest thing is to install it like this:

```bash
go install github.com/bwebb-hx/hxutil
```

Once it's installed this way, you can just run the command in your terminal with `hxutil`.

### Running locally

If you'd like to check out the source code and run it locally, make changes, etc:

```bash
# clone from github
git clone github.com/bwebb-hx/hxutil.git
cd hxutil

# build and run when you wanna test changes
go build
./hxutil <commands, flags, etc>

# to install your local version and use as a normal cli
go install
hxutil <commands, flags, etc>
```

## Usage

For usage, refer to [cli docs](./docs/hxutil.md). It will guide you through the different commands.

## Completion

you can run the following command to generate completions for this cli tool. Make sure to specify the right shell, based on whatever you use for your terminal.

```
hxutil completion [shell type]
```

Here's the full help text from the cli:

```
Generate the autocompletion script for hxutil for the specified shell.
See each sub-command's help for details on how to use the generated script.

Usage:
  hxutil completion [command]

Available Commands:
  bash        Generate the autocompletion script for bash
  fish        Generate the autocompletion script for fish
  powershell  Generate the autocompletion script for powershell
  zsh         Generate the autocompletion script for zsh
```

You can use this command to save the generated completions shell script somewhere, and then source it in your shell's profile. Here's an example of what this might look like:

```bash
hxutil completion zsh > ~/completions/hxutil_completion.sh

# now, source it in your .zshrc file (since we are using zsh here)
echo "source ~/completions/hxutil_completion.sh" >> ~/.zshrc

# after this you can restart your terminal
```

Now, you can hit `tab` to see the next available commands.
