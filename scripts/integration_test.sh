#!/usr/bin/env bash

usage() {
  echo "USAGE: integration_test.sh <PROJECT ROOT> [sub-command]"
  echo ""
  echo "sub-comands:"
  echo -e " - ${CLR_START}${BRIGHT};${GREEN}${CLR_END}all${CLR_RESET}: Run the api & html integration tests"
  echo -e " - ${CLR_START}${BRIGHT};${GREEN}${CLR_END}api${CLR_RESET}: Run the api integration tests"
  echo -e " - ${CLR_START}${BRIGHT};${GREEN}${CLR_END}html${CLR_RESET}: Run the html integration tests"
  echo -e " - ${CLR_START}${BRIGHT};${GREEN}${CLR_END}to-branch${CLR_RESET}: Update the integration tests to the current branch for resweave"
  echo -e " - ${CLR_START}${BRIGHT};${GREEN}${CLR_END}to-version <version>${CLR_RESET}: Update the integration tests to the specified version"
  echo ""
  echo "* Note: No subcommand runs ${CLR_START}${ITALICS}${CLR_END}all${CLR_RESET}"

  exit $STATUS_USAGE_USED
}

if [ $# = 0 ]; then
  usage
fi


PROJECT_ROOT=$1
shift
pushd ${PROJECT_ROOT}
PROJECT_ROOT_MODULE="$(grep 'module' ${PROJECT_ROOT}/go.mod | sed -E 's/^module[[:space:]]*//g')"
INTEGRATION_DIR=${PROJECT_ROOT}/testing/integration
HTML_INTEGRATION_DIR=${INTEGRATION_DIR}/html
API_INTEGRATION_DIR=${INTEGRATION_DIR}/api
EXIT_CODE_LINE="exited with code"
EXIT_CODE_PATTERN="^.*exited with code ([[:digit:]]+).*$"

source ${PROJECT_ROOT}/scripts/status.sh
source ${PROJECT_ROOT}/scripts/log.sh


run_test() {
    file_identifier=$1
    GOOS=linux GOARCH=amd64 go build -o bin/server server.go
    GOOS=linux GOARCH=amd64 go test -c -o bin/test ./...
    echo "Post-Build: ls $PWD/bin"
    ls -F bin/
    docker-compose -f compose.yaml up server &
    sleep 5
    LOG_FILE=${PROJECT_ROOT}/.reports/${file_identifier}_test.log
    docker-compose -f compose.yaml up libtest > ${LOG_FILE} 2>&1 
    cat ${LOG_FILE}
    TEST_RESULT=$(cat ${LOG_FILE} | grep "${EXIT_CODE_LINE}" | sed -E "s/${EXIT_CODE_PATTERN}/\1/g")
    echo "TEST_RESULT=${TEST_RESULT}"
    docker-compose down
}

test_html() {
    pushd ${HTML_INTEGRATION_DIR}
    _DIRS=$(ls -1)
    for file in ${_DIRS}
    do
        if [[ -d ${file} ]]; then
            echo "DIRECTORY: ${file}"
            pushd ${file}
            echo "Pre-Build: ls $PWD/bin"
            mkdir -p bin
            ls -F bin/
            run_test "html_${file}"
            if [[ ${TEST_RESULT} != 0 ]]; then
                echo "---- TESTS FAILED ${file} (${TEST_RESULT})----"
                exit  ${TEST_RESULT}
            fi
            popd
        else
          echo "**** SKIPPING ${file} ****"
        fi
    done
}

test_api() {
    echo "******** RUNNING API INTEGRATION TESTS ********"
    pushd ${API_INTEGRATION_DIR}
    _DIRS=$(ls -1)
    for file in ${_DIRS}
    do
        if [[ -d ${file} ]]; then
            echo "DIRECTORY: ${file}"
            pushd ${file}
            echo "Pre-Build: ls $PWD/bin"
            mkdir -p bin
            ls -F bin/
            run_test "html_${file}"
            if [[ ${TEST_RESULT} != 0 ]]; then
                echo "---- TESTS FAILED ${file} ----"
                exit  ${TEST_RESULT}
            fi
            popd
        else
          echo "**** SKIPPING ${file} ****"
        fi
    done
}

update_all_tests_to_branch() {
    while [[ $# > 0 ]]
    do
        TEST_TYPE=$1
        shift
        pushd ${TEST_TYPE}
        _DIRS=$(ls -1)
        for file in ${_DIRS}
        do
            if [[ -d ${file} ]]; then
                pushd ${file}
                _MOD=`grep module go.mod | sed -E "s/^module (.*)$/\1/g"`
                echo "Updating ${_MOD} to use branch ${BRANCH_NAME}"
                go get ${PROJECT_ROOT_MODULE}@${BRANCH_NAME}
                popd
            fi
        done
        popd
    done
}

update_all_tests_to_tag() {
    TAG=$1
    shift
    while [[ $# > 0 ]]
    do
        TEST_TYPE=$1
        shift
        pushd ${TEST_TYPE}
        _DIRS=$(ls -1)
        for file in ${_DIRS}
        do
            if [[ -d ${file} ]]; then
                pushd ${file}
                _MOD=`grep module go.mod | sed -E "s/^module (.*)$/\1/g"`
                echo "Updating ${_MOD} to use tag ${TAG}"
                go get ${PROJECT_ROOT_MODULE}@${TAG}
                popd
            fi
        done
        popd
    done
}

update_to_branch() {
    BRANCH_NAME=`git rev-parse --abbrev-ref HEAD`
    echo "******** Updating Integration Tests to ${CLR_START};${CLR_GREEN}${CLR_END}${BRANCH_NAME}${CLR_RESET} ********"

    update_all_tests_to_branch ${API_INTEGRATION_DIR} ${HTML_INTEGRATION_DIR}
}

update_to_version() {
    TAG=$1
    echo "******** Updating Integration Tests to ${CLR_START};${CLR_GREEN}${CLR_END}${TAG}${CLR_RESET} ********"

    update_all_tests_to_tag ${TAG} ${API_INTEGRATION_DIR} ${HTML_INTEGRATION_DIR}
}

test_all() {
    test_html
    test_api
}

if [ $# = 0 ]; then
  test_all
else
    case $1 in
      all)
        echo "Running ALL tests"
        test_all
        ;;
      api)
        echo "Running API tests"
        test_api
        ;;
      html)
        echo "Running HTML tests"
        test_html
        ;;
      to-branch)
        echo "Updating tests"
        update_to_branch
        ;;
      to-version)
        echo "Updating test versions to $2"
        update_to_version $2
        ;;
      *)
        usage
        ;;
    esac
fi


