#!/bin/bash

usage() {
  echo "USAGE: integration_test.sh <PROJECT ROOT> [sub-command]"
  echo ""
  echo "sub-comands:"
  echo -e " - ${CLR_START}${BRIGHT};${GREEN}${CLR_END}html${CLR_RESET}: Run the html integration tests"
  echo ""
  echo "* Note: No subcommand runs ${CLR_START}${ITALICS}${CLR_END}html${CLR_RESET}"

  exit $STATUS_USAGE_USED
}

if [ $# = 0 ]; then
  usage
fi


PROJECT_ROOT=$1
shift
pushd ${PROJECT_ROOT}
MODULE_NAME="$(grep 'module' go.mod | sed -E 's/^module[[:space:]]*//g')"
INTEGRATION_DIR=${PROJECT_ROOT}/testing/integration
HTML_INTEGRATION_DIR=${INTEGRATION_DIR}/html

source ${PROJECT_ROOT}/scripts/status.sh
source ${PROJECT_ROOT}/scripts/log.sh

test_html() {
    pushd ${HTML_INTEGRATION_DIR}
    _DIRS=$(ls -1)
    for file in ${_DIRS}
    do
        if [[ -d ${file} ]]; then
            echo "DIRECTORY: ${file}"
            pushd ${file}
            docker-compose up ${file} &
            sleep 1
            docker-compose up libtest
            TEST_RESULT=$?
            docker-compose down
            if [[ ${TEST_RESULT} != 0 ]]; then
                exit  ${TEST_RESULT}
            fi
        else
            echo "SKIPPING: ${file}"
        fi
    done
}


if [ $# = 0 ]; then
  test_html
else
    case $1 in
      html)
        test_html
        ;;
      *)
        usage
        ;;
    esac
fi


