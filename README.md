# tmux-term-drawer

In tmux, toggles a terminal drawer on and off.

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

## Building

The `tmux` dependency is in a private repo, [jplein/tmux](https://github.com/jplein/tmux). You'll need to mark the repository private in order to pull it down, by running a command like this:

```
export GOPRIVATE=github.com/jplein
```