#!/usr/bin/env bash

_docs_script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${_docs_script_dir}/colors.sh

_docs_default_out_dir_=docs/
_docs_default_src_dir_=documentation/

usage() {
  echo "USAGE: compile-docs.sh [sub-command] [OPTIONS]"
  echo ""
  echo "WHERE [sub-command] is one of: "
  echo " - $(color -bright -green all)"
  echo " - $(color -bright -green html)"
  echo " - $(color -bright -green pdf)"
  echo ""
  echo "and OPTIONS can be any of:"
  echo ""
  echo " -o <dirname>: Where to output the build files for all builds (default: ${_docs_default_out_dir_})"
  echo " -r <dirname>: The root directory for the project"
  echo " -s <dirname>: Source directory to build asciidoctor documentation from (default: ${_docs_default_out_dir_})"
  echo ""
  echo " -h: Print this message"

  exit $STATUS_USAGE_USED
}

if [ $# = 0 ]; then
  usage
fi

docs_html() {
  local DOCS_DIR=$1
  shift
  local DOCS_OUT_DIR=$1
  shift

  docker run -u $(id -u):$(id -g) -v ${DOCS_DIR}:/documents/ -v ${DOCS_OUT_DIR}:/docsout/ asciidoctor/docker-asciidoctor asciidoctor -b html5 index.adoc -D /docsout/
}

docs_all() {
  docker image pull asciidoctor/docker-asciidoctor
  docs_html $1 $2
}

run() {
  local _docs_cmd_=$1
  local _docs_out_dir_=${_docs_default_out_dir_}
  local _docs_src_dir_=${_docs_default_src_dir_}
  shift

  OPTIND=1
  while getopts "o:r:s:h" arg; do
    case "${arg}" in
      h)
        usage
        ;;
      r)
        PROJECT_ROOT=${OPTARG}
        ;;
      o)
        _docs_out_dir_=${OPTARG}
        ;;
      s)
        _docs_src_dir_=${OPTARG}
        ;;
      *)
        usage # can modify this in the future to form a _rem_args array which gets passed to sub commands
        ;;
    esac
  done

  if [[ -z "${PROJECT_ROOT}" ]]; then
    PROJECT_ROOT=`git rev-parse --show-toplevel`
  fi

  type docs_${_docs_cmd_} > /dev/null 2>&1
  if [[ $? -ne 0 ]]; then
    echo ""
    echo "Unknown command: '${_docs_cmd_}'"
    echo ""
    usage
  fi

  docs_${_docs_cmd_} ${PROJECT_ROOT}/${_docs_src_dir_} ${PROJECT_ROOT}/${_docs_out_dir_}
}

run $@
