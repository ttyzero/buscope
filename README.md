# buscope

A visual inspector for [ttybus](https://github.com/ttyzero/ttybus).

Sit it next to gitwing (or anything else on the bus). It discovers
channels from connected peers, tails their traffic, and lets you publish
a test line to the selected channel.

```
╭ buscope ●              12 ╮
│ 2  nvim · gitwing         │
│ ▁▂▃▅▇▅▃▂▁▂▃▁▂▁            │
│───────────────────────────│
│ ▸ files     9  ▁▂▃▅▇▃▂    │
│   cwd       1  ▁▁▁▂▁▁▁    │
│───────────────────────────│
│ files                     │
│ 12:01:02  /code/ttybus/a  │
│ 12:01:03  /code/navehnet  │
│ ▸ i pub · files           │
│ j/k  i pub  c clr  q      │
╰───────────────────────────╯
```

Same Charm language as gitwing: `$TTYTHEME`, no panel fill. `--borderless`
(or `$TTYBORDERLESS=1`) drops the rounded frame.
The header dot flashes on traffic. Each channel row has a count and a
sparkline of the last few seconds. Selecting a channel opens its live
log underneath.

## Install

Go 1.25+. `ttybus` on `PATH` (auto-starts the daemon).

```sh
cd ~/code/buscope
make build
./buscope
```

## Use it

```sh
tmux split-window -h ./buscope
```

By default it always watches `files` (gitwing / nvim). Other channels
appear as soon as a peer `SUB`s them (`ttybus ls`).

```sh
buscope --watch files,cwd,editor.open
```

Keys: `j`/`k` select channel, `i` then type to publish, `enter` send,
`esc` back, `c` clear the log, `r` refresh peers, `q` quit.

## Theme

Honors the same stack as gitwing:

```
--theme  >  $TTYTHEME  >  $CLITHEME  >  OSC 11  >  $COLORFGBG  >  dark
--borderless / $TTYBORDERLESS=1
```

Live, on the bus (all three panes):

```sh
ttybus pub theme 'THEME nord'
ttybus pub theme 'THEME catppuccin BORDERS=0'
ttybus pub theme 'BORDERS=0'
```
