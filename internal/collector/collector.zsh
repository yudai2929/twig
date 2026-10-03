autoload -Uz compinit
compinit -D
PROMPT='TWIG> '
RPROMPT=''

# The wrapper observes calls made by the installed zsh completion functions.
# It forwards every call unchanged, so Tab continues to use zsh's own engine.
function compadd() {
  local -a _twig_args _twig_captured_values _twig_captured_descriptions
  _twig_args=("$@")
  local _twig_group='' _twig_values_name='' _twig_descriptions_name='' _twig_suffix=''
  local _twig_i=1
  while (( _twig_i <= $#_twig_args )); do
    case "${_twig_args[_twig_i]}" in
      -a|-A|-O)
        if [[ "${_twig_args[_twig_i]}" == -a ]]; then _twig_values_name="${_twig_args[_twig_i+1]}"; fi
        (( _twig_i += 2 ))
        ;;
      -d|-ld)
        _twig_descriptions_name="${_twig_args[_twig_i+1]}"
        (( _twig_i += 2 ))
        ;;
      -J|-V)
        _twig_group="${_twig_args[_twig_i+1]}"
        (( _twig_i += 2 ))
        ;;
      -S|-s)
        _twig_suffix="${_twig_args[_twig_i+1]}"
        (( _twig_i += 2 ))
        ;;
      *) (( _twig_i++ )) ;;
    esac
  done

  builtin compadd "$@"
  local _twig_result=$?
  [[ -n "$_twig_values_name" ]] || return $_twig_result
  _twig_captured_values=( "${(@P)_twig_values_name}" )
  if [[ -n "$_twig_descriptions_name" ]]; then
    _twig_captured_descriptions=( "${(@P)_twig_descriptions_name}" )
  fi
  for (( _twig_i=1; _twig_i <= $#_twig_captured_values; _twig_i++ )); do
    printf '%s\0%s\0%s\0%s\0%s\0' "$_twig_group" "${_twig_captured_values[_twig_i]}" "${_twig_captured_descriptions[_twig_i]}" "$_twig_suffix" '' >> "$TWIG_RESULT_FILE"
  done
  return $_twig_result
}

function twig_generated_completion() {
  local command_name="$1" help_text script_text script_file
  (( $+commands[$command_name] )) || return 1
  help_text=$("$command_name" --help 2>/dev/null)
  [[ "${(L)help_text}" == *completion* ]] || return 1
  script_text=$("$command_name" completion zsh 2>/dev/null)
  if [[ "$script_text" != '#compdef '* ]]; then
    script_text=$("$command_name" completion -s zsh 2>/dev/null)
  fi
  [[ "$script_text" == '#compdef '* && ${#script_text} -lt 1000000 ]] || return 1
  script_file=$(mktemp "${TMPDIR:-/tmp}/twig-generated.XXXXXXXX") || return 1
  print -r -- "$script_text" > "$script_file"
  source "$script_file"
  local result=$?
  rm -f -- "$script_file"
  return $result
}

function twig_capture_widget() {
  BUFFER=$(<"$TWIG_BUFFER_FILE")
  CURSOR=$TWIG_CURSOR
  local command_name="${BUFFER%%[[:space:]]*}"
  if [[ -z "${_comps[$command_name]-}" ]]; then
    twig_generated_completion "$command_name"
  fi
  zle list-choices
  : > "$TWIG_DONE_FILE"
  zle -I
}
zle -N twig_capture_widget
bindkey '^Xg' twig_capture_widget
print -r -- TWIG_READY
