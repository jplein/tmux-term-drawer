# tmux-term-drawer

In tmux, shows or hides a terminal pane. Running it once creates and shows the terminal pane, running it again hides it, and so on. The anticipated use case is using it with an IDE-ish Neovim or Emacs/emacsclient instance filling up your active tmux pane, and you use this tool to hide and show a terminal window to do things like commit or push your changes, trigger build or linting steps, etc.

The state of your terminal pane is retained after you hide and show the pane. When hidden, it's placed into a different tmux session named "term-drawer." If you start a long-lived process like a build, it will keep running even if the pane is hidden.

Each tmux window can have its own terminal pane, and each is completely independent of the others. (The tool does not create a terminal pane for a tmux window until it has been run in that tmux window.)

By default, the terminal pane is placed on the right side of the tmux window, taking up 30% of the width of the active window, but the size and position can be customized (see below).

After you've installed it and put it into your path, you'll probably want to configure a keyboard shortcut for it in your `tmux.conf`. This is what I'm using:

```
bind-key 't' run-shell tmux-term-drawer
```

If you're using the default tmux prefix of `control-b`, you would type `control-b`, followed by `t`.

## Requirements

Requires tmux 3.3a or 3.4. It may work with other versions, but I've received a report of it failing with tmux 3.1, and it has not been tested with 3.2 or 3.3.

It is known to work macOS, Debian, and Ubuntu. I have not tested it on Windows, and do not have a Windows install to test it on.

## Installation

Clone this repo and, from the repo directory, run this command:

```
go build && go install
```

## Configuration

Create a file in the root of your home directory named `.term-drawer-config.yaml` like this:

```
tmux-term-drawer:
    # Where to put the drawer. One of top, bottom, left, right.
    position: right

    # The size of the drawer. If units is 'percent', the height or width of the
    # drawer is a percentage of the active window's height or width. If units is
    # 'character', the drawer is a fixed number of characters.
    size: 30
    units: percent

    # You will generally not need to change this. This is the name of the tmux
    # session which holds the drawer when it is hidden
    sessionName: term-drawer
```
