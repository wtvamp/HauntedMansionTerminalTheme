#!/usr/bin/env zsh
# Loads the generated tool themes (bat, eza, fzf, starship, less).
#
# The generated file does the work; this exists so ~/.zshrc.d has a stable
# filename to symlink and so HM_ROOT is resolved the same way everywhere.

() {
  emulate -L zsh -o null_glob
  local root=${HM_ROOT:-${${(%):-%x}:A:h:h}}
  [[ -r $root/dist/shell/_tools.zsh ]] && source $root/dist/shell/_tools.zsh
}
