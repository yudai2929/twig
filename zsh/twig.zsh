# Source this file from an interactive zsh after building cmd/twig.
[[ -o interactive ]] || return 0
(( ${_twig_active:-0} )) && return 0
autoload -Uz add-zle-hook-widget
autoload -Uz add-zsh-hook
if (( ! $+functions[compdef] )); then
  autoload -Uz compinit
  compinit -D
fi

typeset -g TWIG_BIN="${TWIG_BIN:-twig}"
typeset -gi _twig_active=1
typeset -gA _twig_previous_widgets
typeset -ga _twig_bound_keys
typeset -gA _twig_files _twig_revisions
typeset -ga _twig_values _twig_descriptions _twig_suffixes _twig_kinds
typeset -gi _twig_revision=${_twig_revision:-0} _twig_index=1 _twig_suppressed=0 _twig_skip_next_space=0 _twig_last_cursor=-1
(( _twig_revision++ ))
typeset -g _twig_last_buffer=''
typeset -g _twig_context=''
typeset -g _twig_prefix_cached=''
typeset -g _twig_rendered=''
typeset -gi _twig_render_changed=0
typeset -g _twig_generation_file
_twig_generation_file=$(mktemp "${TMPDIR:-/tmp}/twig-generation.XXXXXXXX") || return 1

function _twig_prefix() {
  local left="$LBUFFER"
  while [[ -n "$left" && "${left[-1]}" != ' ' && "${left[-1]}" != $'\t' && "${left[-1]}" != '/' ]]; do
    left="${left[1,-2]}"
  done
  REPLY="${LBUFFER#$left}"
}

function _twig_word() {
  local left="$LBUFFER"
  while [[ -n "$left" && "${left[-1]}" != ' ' && "${left[-1]}" != $'\t' && "${left[-1]}" != $'\n' ]]; do
    left="${left[1,-2]}"
  done
  REPLY="${LBUFFER#$left}"
}

function _twig_exact_command() {
  [[ -n "$BUFFER" && "$BUFFER" == "$LBUFFER" && "$BUFFER" != *[[:space:]]* ]] && (( $+commands[$BUFFER] ))
}

function _twig_draw() {
  if (( $#_twig_values == 0 || _twig_suppressed )); then
    _twig_render_changed=0
    if [[ -n "$_twig_rendered" ]]; then
      _twig_rendered=''
      _twig_render_changed=1
      zle -M ''
    fi
    return
  fi
  local text='' i marker description icon
  local first=$(( _twig_index > 8 ? _twig_index - 7 : 1 ))
  local last=$(( first + 7 < $#_twig_values ? first + 7 : $#_twig_values ))
  for (( i=first; i <= last; i++ )); do
    marker='  '
    (( i == _twig_index )) && marker='› '
    description="${_twig_descriptions[i]}"
    [[ "$description" == *' -- '* ]] && description="${description#* -- }"
    case "${_twig_kinds[i]}" in
      command) icon='⚙' ;;
      flag) icon='⚑' ;;
      file) icon='📄' ;;
      directory) icon='📁' ;;
      *) icon='•' ;;
    esac
    text+="${icon} ${marker}${_twig_values[i]}"
    [[ -n "$description" ]] && text+="  ${description}"
    (( i < last )) && text+=$'\n'
  done
  _twig_render_changed=0
  if [[ "$text" != "$_twig_rendered" ]]; then
    _twig_rendered="$text"
    _twig_render_changed=1
    zle -M "$text"
  fi
}

function _twig_result_ready() {
  local fd=$1 line file revision
  IFS= read -r line <&$fd
  zle -F "$fd"
  exec {fd}<&-
  file="${_twig_files[$fd]}"
  revision="${_twig_revisions[$fd]}"
  unset "_twig_files[$fd]" "_twig_revisions[$fd]"
  if [[ "$revision" != "$_twig_revision" || $_twig_suppressed -eq 1 || $_twig_active -eq 0 ]]; then
    rm -f -- "$file"
    return
  fi
  _twig_values=()
  _twig_descriptions=()
  _twig_suffixes=()
  _twig_kinds=()
  _twig_index=1
  if [[ -s "$file" ]]; then
    local result_fd group value description suffix kind
    exec {result_fd}< "$file"
    while IFS= read -r -d '' group <&$result_fd; do
      IFS= read -r -d '' value <&$result_fd || break
      IFS= read -r -d '' description <&$result_fd || break
      IFS= read -r -d '' suffix <&$result_fd || break
      IFS= read -r -d '' kind <&$result_fd || break
      _twig_values+=("$value")
      _twig_descriptions+=("$description")
      _twig_suffixes+=("$suffix")
      _twig_kinds+=("$kind")
    done
    exec {result_fd}<&-
  fi
  rm -f -- "$file"
  _twig_prefix
  _twig_context="${LBUFFER[1,$(( ${#LBUFFER} - ${#REPLY} ))]}"
  _twig_prefix_cached="$REPLY"
  _twig_draw
  (( _twig_render_changed )) && zle -R
}

function _twig_start_request() {
  local result_file fd revision="$_twig_revision"
  result_file=$(mktemp "${TMPDIR:-/tmp}/twig-result.XXXXXXXX") || return
  exec {fd}< <(sleep 0.06; if [[ "$(<$_twig_generation_file)" == "$revision" ]]; then FPATH="${(j.:.)fpath}" "$TWIG_BIN" complete --buffer "$BUFFER" --cursor "$CURSOR" --cwd "$PWD" > "$result_file" 2>/dev/null; fi; print -r -- ready)
  _twig_files[$fd]="$result_file"
  _twig_revisions[$fd]="$_twig_revision"
  zle -F "$fd" _twig_result_ready
}

function _twig_on_redraw() {
  if [[ "$BUFFER" == "$_twig_last_buffer" && "$CURSOR" == "$_twig_last_cursor" ]]; then
    return
  fi
  _twig_last_buffer="$BUFFER"
  _twig_last_cursor=$CURSOR
  _twig_skip_next_space=0
  (( _twig_revision++ ))
  print -r -- "$_twig_revision" > "$_twig_generation_file"
  _twig_suppressed=0
  if [[ -z "$BUFFER" ]] || _twig_exact_command; then
    _twig_values=()
    _twig_descriptions=()
    _twig_suffixes=()
    _twig_kinds=()
    _twig_index=1
    _twig_draw
    return
  fi
  _twig_prefix
  local prefix="$REPLY" context="${LBUFFER[1,$(( ${#LBUFFER} - ${#REPLY} ))]}" i
  _twig_word
  local word="$REPLY"
  if (( $#_twig_values )) && [[ "$context" == "$_twig_context" && "$prefix" == "$_twig_prefix_cached"* ]]; then
    local old_count=$#_twig_values old_index=$_twig_index
    local -a values descriptions suffixes kinds
    for (( i=1; i <= $#_twig_values; i++ )); do
      if [[ "${_twig_values[i]}" == "$prefix"* || "${_twig_values[i]}" == "$word"* ]]; then
        values+=("${_twig_values[i]}")
        descriptions+=("${_twig_descriptions[i]}")
        suffixes+=("${_twig_suffixes[i]}")
        kinds+=("${_twig_kinds[i]}")
      fi
    done
    _twig_values=("${values[@]}")
    _twig_descriptions=("${descriptions[@]}")
    _twig_suffixes=("${suffixes[@]}")
    _twig_kinds=("${kinds[@]}")
    _twig_index=1
    _twig_prefix_cached="$prefix"
    if (( $#_twig_values != old_count || old_index != 1 )); then
      _twig_draw
    fi
    (( $#_twig_values )) && return
  else
    _twig_values=()
    _twig_descriptions=()
    _twig_suffixes=()
    _twig_kinds=()
    _twig_index=1
    _twig_draw
  fi
  _twig_start_request
}

function _twig_accept() {
  if _twig_exact_command; then
    zle .accept-line
    return
  fi
  if (( $#_twig_values == 0 || _twig_suppressed )); then
    zle .accept-line
    return
  fi
  _twig_prefix
  local prefix="$REPLY" value="${_twig_values[_twig_index]}" suffix="${_twig_suffixes[_twig_index]}" kind="${_twig_kinds[_twig_index]}"
  _twig_word
  [[ "$value" == "$REPLY"* ]] && prefix="$REPLY"
  local keep=$(( ${#LBUFFER} - ${#prefix} ))
  local left="${LBUFFER[1,$keep]}"
  local advance=0
  [[ -z "$RBUFFER" && "$suffix" == ' ' && "$kind" == command ]] && advance=1
  LBUFFER="${left}${(q)value}${suffix}"
  _twig_suppressed=$(( ! advance ))
  _twig_skip_next_space=0
  [[ "$suffix" == ' ' ]] && _twig_skip_next_space=1
  (( _twig_revision++ ))
  print -r -- "$_twig_revision" > "$_twig_generation_file"
  _twig_values=()
  _twig_kinds=()
  _twig_draw
  _twig_last_buffer="$BUFFER"
  _twig_last_cursor=$CURSOR
  zle -R
  (( advance )) && _twig_start_request
}

function _twig_space() {
  if (( _twig_skip_next_space )) && [[ "${LBUFFER[-1]}" == ' ' ]]; then
    _twig_skip_next_space=0
    if (( _twig_suppressed )); then
      _twig_last_buffer=''
      _twig_on_redraw
    fi
    return
  fi
  if (( _twig_suppressed )); then
    _twig_suppressed=0
    [[ "${LBUFFER[-1]}" == ' ' ]] || zle .self-insert
    _twig_last_buffer=''
    _twig_on_redraw
    return
  fi
  zle .self-insert
}

function _twig_after_history() {
  _twig_values=()
  _twig_descriptions=()
  _twig_suffixes=()
  _twig_kinds=()
  _twig_index=1
  _twig_suppressed=1
  _twig_skip_next_space=0
  _twig_last_buffer="$BUFFER"
  _twig_last_cursor=$CURSOR
  (( _twig_revision++ ))
  print -r -- "$_twig_revision" > "$_twig_generation_file"
  _twig_draw
}

function _twig_up() {
  if (( $#_twig_values && ! _twig_suppressed )); then
    (( _twig_index > 1 )) && (( _twig_index-- ))
    _twig_draw
    return
  fi
  local previous_buffer="$BUFFER"
  zle .up-line-or-history
  [[ "$BUFFER" != "$previous_buffer" ]] && _twig_after_history
}

function _twig_down() {
  if (( $#_twig_values && ! _twig_suppressed )); then
    (( _twig_index < $#_twig_values )) && (( _twig_index++ ))
    _twig_draw
    return
  fi
  local previous_buffer="$BUFFER"
  zle .down-line-or-history
  [[ "$BUFFER" != "$previous_buffer" ]] && _twig_after_history
}

function _twig_escape() {
  _twig_values=()
  _twig_kinds=()
  _twig_suppressed=1
  (( _twig_revision++ ))
  print -r -- "$_twig_revision" > "$_twig_generation_file"
  _twig_draw
}

function _twig_tab() {
  _twig_values=()
  _twig_kinds=()
  _twig_suppressed=0
  _twig_skip_next_space=0
  (( _twig_revision++ ))
  print -r -- "$_twig_revision" > "$_twig_generation_file"
  zle expand-or-complete
}

function _twig_reset() {
  _twig_values=()
  _twig_kinds=()
  _twig_suppressed=0
  _twig_skip_next_space=0
  _twig_last_buffer=''
  _twig_context=''
  _twig_prefix_cached=''
  _twig_last_cursor=-1
  _twig_rendered=''
  (( _twig_revision++ ))
  print -r -- "$_twig_revision" > "$_twig_generation_file"
}

function _twig_cleanup() {
  rm -f -- "$_twig_generation_file" "${(@v)_twig_files}"
}

function _twig_unload() {
  (( _twig_active )) || return 0
  _twig_active=0
  add-zle-hook-widget -d line-pre-redraw _twig_on_redraw
  add-zle-hook-widget -d line-init _twig_reset
  add-zsh-hook -d zshexit _twig_cleanup
  local key
  for key in "${_twig_bound_keys[@]}"; do
    bindkey -- "$key" "${_twig_previous_widgets[$key]}"
  done
  _twig_cleanup
}

zle -N _twig_accept
zle -N _twig_space
zle -N _twig_up
zle -N _twig_down
zle -N _twig_escape
zle -N _twig_tab
_twig_bound_keys=('^M' ' ' '^Xj' '^Xk' '^[[A' '^[[B' '^[OA' '^[OB' '^[' '^I')
typeset _twig_key _twig_binding
for _twig_key in "${_twig_bound_keys[@]}"; do
  _twig_binding=$(bindkey -- "$_twig_key")
  _twig_previous_widgets[$_twig_key]="${_twig_binding##* }"
done
bindkey '^M' _twig_accept
bindkey ' ' _twig_space
bindkey '^Xk' _twig_up
bindkey '^Xj' _twig_down
bindkey '^[[A' _twig_up
bindkey '^[[B' _twig_down
bindkey '^[OA' _twig_up
bindkey '^[OB' _twig_down
bindkey '^[' _twig_escape
bindkey '^I' _twig_tab
add-zle-hook-widget line-pre-redraw _twig_on_redraw
add-zle-hook-widget line-init _twig_reset
add-zsh-hook zshexit _twig_cleanup
