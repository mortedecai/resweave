#!/usr/bin/env bash

_lint_script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${_lint_script_dir}/colors.sh

usage() {
  echo "USAGE: lint.sh <PROJECT ROOT> [sub-command]"
  echo ""
  echo "sub-commands:"
  echo " - $(color -bright -green all)"
  echo ""
  echo "* Note: No subcommand runs $(color -italics all)"

  exit $STATUS_USAGE_USED
}

if [ $# = 0 ]; then
  usage
fi


PROJECT_ROOT=$1
shift

source ${PROJECT_ROOT}/scripts/status.sh
source ${PROJECT_ROOT}/scripts/log.sh

check_for_staticcheck() {
  which staticcheck
  if [ $? != 0 ]; then
    create_error_log_line "staticcheck not found."
    exit $STATUS_APP_NOT_FOUND
  fi
}

check_binaries() {
  check_for_staticcheck
}

run_staticcheck() {
  fmt="stylish"
  if [[ $# = 1 ]]; then
    fmt="$1"
  fi

  pushd ${PROJECT_ROOT}
  staticcheck -f ${fmt}
  popd
}
lint_all() {
  check_binaries

  run_staticcheck
}

lint_all_json() {
  check_binaries

  run_staticcheck "json"
}

if [ $# = 0 ]; then
  lint_all
else
    case $1 in
      all)
        lint_all
        ;;
      staticcheck)
        staticcheck
        ;;
      cicd)
        lint_all_json
        ;;
      *)
        usage
        ;;
    esac
fi


