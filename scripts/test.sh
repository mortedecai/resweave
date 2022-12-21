#!/bin/bash

usage() {
  echo "USAGE: test.sh <PROJECT ROOT> [sub-command]"
  echo ""
  echo "sub-comands:"
  echo -e " - ${CLR_START}${BRIGHT};${GREEN}${CLR_END}all${CLR_RESET}"
  echo ""
  echo "* Note: No subcommand runs ${CLR_START}${ITALICS}${CLR_END}all${CLR_RESET}"

  exit $STATUS_USAGE_USED
}

if [ $# = 0 ]; then
  usage
fi


PROJECT_ROOT=$1
shift

source ${PROJECT_ROOT}/scripts/status.sh
source ${PROJECT_ROOT}/scripts/log.sh

test_unit() {
    go test $(grep "module" go.mod | sed -E "s/^module[[:space:]]*//g")
}

test_integration() {
  ${PROJECT_ROOT}/scripts/integration_test.sh $PROJECT_ROOT $@

  echo "++++ Integration Tests Passed ++++"
}

test_all() {
    pushd ${PROJECT_ROOT}
    test_unit
    test_integration
}


if [ $# = 0 ]; then
  test_all
else
    case $1 in
      all)
        test_all
        ;;
      unit)
        test_unit
        ;;
      integration)
        test_integration
        ;;
      *)
        usage
        ;;
    esac
fi


