#!/usr/bin/env zsh
# Haunted Mansion greeting: one ghost, one quote, on every new terminal.
#
# LATENCY BUDGET: 30ms. This runs on every shell Warren opens, and he runs 20-30
# agents at once, so the cost is multiplied by that. Everything below is a zsh
# builtin on purpose:
#
#   - $(<file) is a builtin read in zsh, NOT a fork like it is in bash
#   - ${(f)...} splits on newlines without calling out to anything
#   - ${array:#pattern} filters without grep
#   - $RANDOM indexes without shuf, sort -R, awk or head
#
# Do not add: node, python, bin/doombuggy, curl, or a pipeline. If you need
# something this cannot express, it belongs in the ride, not the greeting.
# Measure before and after:  hyperfine --warmup 3 'zsh -ic true'

() {
  emulate -L zsh -o no_aliases -o extended_glob

  local root=${HM_ROOT:-${${(%):-%x}:A:h:h}}
  local colors=$root/dist/shell/_colors.zsh
  [[ -r $colors ]] && source $colors

  # Roles fall back to plain slot numbers so the greeting still works if the
  # generated colors are missing (a fresh clone before `make generate`).
  local quote_color=${HM_ROLE_GREETING_QUOTE:-8}

  # The portraits are photographs rendered as half-block glyphs, so they only
  # exist in colour -- there is no meaningful monochrome form of a photograph.
  # Under NO_COLOR the greeting is the quote alone.
  local -a ghosts
  if [[ -z $NO_COLOR && -d $root/dist/ghosts ]]; then
    ghosts=( $root/dist/ghosts/*.ans(N) )
  fi

  local -a quotes
  quotes=( ${(f)"$(<$root/content/quotes.txt)"} )
  quotes=( ${quotes:#(\#*|[[:space:]]#)} )
  (( $#quotes )) || return 0

  local quote=${quotes[RANDOM % $#quotes + 1]}

  # Raw SGR rather than print -P: the art contains % and backslashes that prompt
  # expansion would eat, and one escape either side is cheaper than escaping it.
  local dim=$'\e[38;5;'${quote_color}m off=$'\e[0m'

  print -r -- ""
  if (( $#ghosts )); then
    print -r -- "$(<${ghosts[RANDOM % $#ghosts + 1]})"
    print -r -- "  ${dim}${quote}${off}"
  else
    print -r -- "  ${quote}"
  fi
  print -r -- ""
}
