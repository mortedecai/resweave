#!/usr/bin/env bash

_cvg_script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${_cvg_script_dir}/colors.sh

usage() {
  echo "USAGE: coverage.sh [report-types]"
  echo ""
  echo "report-types:"
  echo " - $(color -bright -green all)"
  echo " - $(color -bright -green html)"
  echo " - $(color -bright -green lcov)"

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
  go test $1 $(grep "module" ${PROJECT_ROOT}/go.mod | sed -E "s/^module[[:space:]]*//g")
}

generate_coverage_file() {
  pushd ${PROJECT_ROOT}

  if [[ ! -e "${REPORTS_DIR}" ]]; then
      mkdir -p ${REPORTS_DIR}
  fi

  echo "Coverage profile location: ${REPORTS_DIR}/coverage.out"
  generate_coverage "--coverprofile=${REPORTS_DIR}/coverage.out"
}

generate_coverage_report() {
  local TYPE=$1
  go tool cover -${TYPE}=${REPORTS_DIR}/coverage.out -o ${REPORTS_DIR}/coverage.html
}

coverage_lcov() {
  generate_coverage_file
}

coverage_html() {
  coverage_lcov
  generate_coverage_report html
}

coverage_all() {
  coverage_html
}

if [ $# = 0 ]; then
  coverage_all
else
    case $1 in
      lcov)
        coverage_lcov
        ;;
      html)
        coverage_html
        ;;
      all)
        coverage_all
        ;;
      *)
        usage
        ;;
    esac
fi



