#!/usr/bin/env zsh
# Haunted Mansion prompt.
#
# Colors come from dist/shell/_colors.zsh via role names — this file never spells
# out a color, so re-running `make generate` after a palette change updates the
# prompt for free. If you find yourself wanting a color that has no role, add the
# role to palette/haunted-mansion.yaml rather than a literal here.

# Starship, if installed, is the better prompt: it shows git state, language
# versions and cloud context, and it is configured from the same palette
# (dist/starship/). This file is the zero-dependency fallback, so it gets out of
# the way rather than fighting starship for $PROMPT.
if (( $+commands[starship] )) && [[ -n ${STARSHIP_CONFIG:-} ]]; then
  return 0
fi

() {
  emulate -L zsh
  local root=${HM_ROOT:-${${(%):-%x}:A:h:h}}
  [[ -r $root/dist/shell/_colors.zsh ]] && source $root/dist/shell/_colors.zsh
}

# Fallbacks keep the prompt sane on a fresh clone, before `make generate` has run.
: ${HM_ROLE_PROMPT_PATH:=6}
: ${HM_ROLE_PROMPT_GIT_CLEAN:=2}
: ${HM_ROLE_PROMPT_GIT_DIRTY:=3}
: ${HM_ROLE_PROMPT_GIT_AHEAD:=5}
: ${HM_ROLE_PROMPT_ERROR:=1}
: ${HM_ROLE_PROMPT_SYMBOL:=5}

autoload -Uz vcs_info
zstyle ':vcs_info:*' enable git
zstyle ':vcs_info:*' check-for-changes true
# Deliberately terse: check-for-changes already costs a git call, and staged/
# unstaged detail past "something is dirty" is not worth a second one.
zstyle ':vcs_info:git:*' unstagedstr  '✦'
zstyle ':vcs_info:git:*' stagedstr    '✧'
zstyle ':vcs_info:git:*' formats      '%b%u%c'
zstyle ':vcs_info:git:*' actionformats '%b|%a%u%c'

_hm_vcs() {
  vcs_info
  [[ -z $vcs_info_msg_0_ ]] && return

  # Dirty if vcs_info appended either marker.
  local color=$HM_ROLE_PROMPT_GIT_CLEAN
  [[ $vcs_info_msg_0_ == *[✦✧]* ]] && color=$HM_ROLE_PROMPT_GIT_DIRTY

  local ahead
  local -a counts
  counts=( ${(f)"$(git rev-list --left-right --count @{upstream}...HEAD 2>/dev/null)"} )
  if [[ -n $counts ]]; then
    local behind_n=${${(z)counts}[1]} ahead_n=${${(z)counts}[2]}
    (( ahead_n > 0 )) && ahead+="%F{$HM_ROLE_PROMPT_GIT_AHEAD}↑$ahead_n%f"
    (( behind_n > 0 )) && ahead+="%F{$HM_ROLE_PROMPT_GIT_AHEAD}↓$behind_n%f"
  fi

  print -n " %F{$color}⑂ ${vcs_info_msg_0_}%f$ahead"
}

autoload -Uz add-zsh-hook
add-zsh-hook precmd _hm_precmd
_hm_precmd() { _HM_VCS=$(_hm_vcs) }

setopt prompt_subst

# Line 1: where you are, and what the repository thinks of you.
# Line 2: the tombstone you type after.
PROMPT='
 %F{$HM_ROLE_PROMPT_PATH}⚰ %~%f${_HM_VCS}
 %(?.%F{$HM_ROLE_PROMPT_SYMBOL}.%F{$HM_ROLE_PROMPT_ERROR})†%f '

# Exit status only when something died, and the time it died at.
RPROMPT='%(?..%F{$HM_ROLE_PROMPT_ERROR}✝ %?%f  )%F{$HM_ROLE_PROMPT_GIT_AHEAD}%D{%H:%M}%f'
