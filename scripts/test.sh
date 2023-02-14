#!/usr/bin/env bash

usage() {
  echo "USAGE: test.sh <PROJECT ROOT> [sub-command]"
  echo ""
  echo "sub-comands:"
  echo -e " - ${CLR_START}${BRIGHT};${GREEN}${CLR_END}all${CLR_RESET}"
  echo -e " - ${CLR_START}${BRIGHT};${GREEN}${CLR_END}unit${CLR_RESET}"
  echo -e " - ${CLR_START}${BRIGHT};${GREEN}${CLR_END}integration${CLR_RESET}"
  echo ""
  echo "* Note: No subcommand runs ${CLR_START}${ITALICS}${CLR_END}all${CLR_RESET}"

  exit $STATUS_USAGE_USED
}

if [ $# = 0 ]; then
  usage
fi


PROJECT_ROOT=$1
SILENT_TESTS=0
shift
if [[ $0 == "-s" ]]; then
    SILENT_TESTS=1
fi

source ${PROJECT_ROOT}/scripts/status.sh
source ${PROJECT_ROOT}/scripts/log.sh

UNIT_TEST_RESULT=0
INT_TEST_RESULT=0

test_unit() {
    go test $(grep "module" go.mod | sed -E "s/^module[[:space:]]*//g") > ${PROJECT_ROOT}/.reports/unit_tests.log 2>&1
    UNIT_TEST_RESULT=$?
    if [[ $SILENT_TESTS == 0 ]]; then
        cat ${PROJECT_ROOT}/.reports/unit_tests.log
    fi
}

test_integration() {
    ${PROJECT_ROOT}/scripts/integration_test.sh $PROJECT_ROOT $* > ${PROJECT_ROOT}/.reports/integration_tests.log 2>&1
    INT_TEST_RESULT=$?
    if [[ $SILENT_TESTS == 0 ]]; then
        cat ${PROJECT_ROOT}/.reports/integration_tests.log
    fi
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
        shift
        test_all
        ;;
      unit)
        shift
        test_unit
        ;;
      integration)
        shift
        test_integration $*
        ;;
      *)
        usage
        ;;
    esac
fi

if [[ 0 != $UNIT_TEST_RESULT ]]; then
  echo "---- Unit Tests Failed ----"
  if [[ 0 != $INT_TEST_RESULT ]]; then
      echo "---- Integration Tests Failed ----"
      exit $STATUS_ALL_TEST_FAILED
  else
      echo "++++ Integration Tests Passed ++++"
      exit $STATUS_UNIT_TEST_FAILED
  fi
else
  echo "++++ Unit Tests Passed ++++"
  if [[ 0 != $INT_TEST_RESULT ]]; then
      echo "---- Integration Tests Failed ----"
      exit $STATUS_INTEGRATION_TEST_FAILED
  fi
  echo "++++ Integration Tests Passed ++++"
fi

