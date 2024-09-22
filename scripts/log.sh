#!/usr/bin/env bash

_log_script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${_log_script_dir}/colors.sh

create_info_log_line() {
  local _lnHdr="$(color -bright -lt_red ===)"
  echo "${_lnHdr}   $1   ${_lnHdr}"
}

log_info_stage() {
  local _stage="$(color -bright -lt_gray $1 project)"
  local _name="$(color -bright -green $2)"
  create_info_log_line "${_stage} ${_name}"
}

create_error_log_line() {
  echo "$(color -white -bg-red ---\ \ \ $1\ \ \ ---)"
}

