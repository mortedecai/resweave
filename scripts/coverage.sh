#!/usr/bin/env bash

_cvg_script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${_cvg_script_dir}/colors.sh

usage() {
  echo "USAGE: coverage.sh <PROJECT ROOT> [report-types]"
  echo ""
  echo "report-types:"
  echo " - $(color -bright -green all)"
  echo " - $(color -bright -green html)"
  echo " - $(color -bright -green out)"
  echo ""
  echo "* Note: No subcommand runs $(color -italics all)"

  exit $STATUS_USAGE_USED
}

if [ $# = 0 ]; then
  usage
fi


PROJECT_ROOT=$1
shift
REPORTS_DIR=${PROJECT_ROOT}/.reports

source ${PROJECT_ROOT}/scripts/status.sh
source ${PROJECT_ROOT}/scripts/log.sh

generate_coverage() {
    GOPRIVATE=github.com/mortedecai go get -u
    GOPRIVATE=github.com/mortedecai go test $1 $(grep "module" ${PROJECT_ROOT}/go.mod | sed -E "s/^module[[:space:]]*//g")
}

generate_coverage_file() {
  echo "Coverage profile location: ${REPORTS_DIR}/coverage.out"
  generate_coverage "--coverprofile=${REPORTS_DIR}/coverage.out"
}

coverage_all() {
    pushd ${PROJECT_ROOT}

    if [[ ! -e "${REPORTS_DIR}" ]]; then
        mkdir -p ${REPORTS_DIR}
    fi
    generate_coverage_file
}


if [ $# = 0 ]; then
  coverage_all
else
    case $1 in
      all)
        coverage_all
        ;;
      *)
        usage
        ;;
    esac
fi



