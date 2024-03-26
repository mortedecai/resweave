#!/usr/bin/env bash

bump_usage() {
echo "USAGE: bump.sh -r <PROJECT ROOT> [OPTIONS]"
  echo ""
  echo "Where OPTIONS is one or more of:"
  echo ""
  echo "-d Dry run (don't commit tags to git)"
  echo ""
  echo "-c Print the current version and exit"
  echo "-h Print this message"
  echo "-o Omit the alpha portion of the version."
  echo "-r Set the project root directory."
  echo "-s Silent (no output except final output)"
  echo "-v [alpha|patch|minor|major] Bump the appropriate tag [defaults to alpha if not supplied]."
}

if [ $# = 0 ]; then
  bump_usage $@
  exit 1
fi

PROJECT_ROOT=""
BUMP_VERSION="alpha"
OMIT_ALPHA=0
DRY_RUN=0
PRINT_CUR_REV=0
PROJECT_SILENT=0

OPTIND=1
while getopts "cdhor:sv:" arg; do
    case "${arg}" in
      d)
        DRY_RUN=1
        ;;
      h)
        bump_usage $@
        exit 1
        ;;
      o)
        OMIT_ALPHA=1
        ;;
      r)
        PROJECT_ROOT=${OPTARG}
        ;;
      s)
        PROJECT_SILENT=1
        ;;
      c)
        PRINT_CUR_REV=1
        ;;
      v)
        BUMP_VERSION=${OPTARG}
        ;;
    esac
done
CUR_VERSION=$(git describe --tags --abbrev=0 2> /dev/null)

if [ $? -ne 0 ]; then
    CUR_VERSION="v0.0.0"
fi

if [ ${PRINT_CUR_REV} -eq 1 ]; then
    echo ${CUR_VERSION}
    exit 0
fi

if [[ -z ${PROJECT_ROOT} ]]; then
    bump_usage
    exit 5
fi

if [ ${PROJECT_SILENT} -ne 1 ]; then
    echo "Bumping from ${CUR_VERSION}"
fi

MAJ_VERSION=$(echo "${CUR_VERSION}" | sed -E "s/^v?([[:digit:]]+)\..*$/\1/g")
MIN_VERSION=$(echo "${CUR_VERSION}" | sed -E "s/^v?[[:digit:]]+\.([[:digit:]]+)\..*$/\1/g")
PAT_VERSION=$(echo "${CUR_VERSION}" | sed -E "s/^v?[[:digit:]]+\.[[:digit:]]+\.([[:digit:]]+)-?.*$/\1/g")
ALPHA_DATE=$(echo "${CUR_VERSION}" | sed -E "s/^v?[[:digit:]]+\.[[:digit:]]+\.[[:digit:]]+-alpha-([[:digit:]]+)\..*$/\1/g")
ALPHA_VERSION=$(echo "${CUR_VERSION}" | sed -E "s/^v?[[:digit:]]+\.[[:digit:]]+\.[[:digit:]]+-alpha-[[:digit:]]+\.([[:digit:]]+).*$/\1/g")

# Good for debugging
#echo -e "Major Version: ${MAJ_VERSION}\nMinor Version: ${MIN_VERSION}\nPatch Version: ${PAT_VERSION}\nAlpha Version: ${ALPHA_VERSION}\nAlpha Date: ${ALPHA_DATE}\nAlpha Version: ${ALPHA_VERSION}"
#echo -e "Assembled    : ${MAJ_VERSION}.${MIN_VERSION}.${PAT_VERSION}-alpha-${ALPHA_DATE}.${ALPHA_VERSION}"

case ${BUMP_VERSION} in
  alpha)
    ALPHA_NEW_DATE=$(date '+%m%d%y')
    if [[ ${ALPHA_NEW_DATE} != ${ALPHA_DATE} ]]; then
        ALPHA_VERSION=1
    else
        ALPHA_VERSION=$((ALPHA_VERSION+1))
    fi
    ALPHA_DATE=${ALPHA_NEW_DATE}
    ;;
  major)
    MAJ_VERSION=$((MAJ_VERSION+1))
    MIN_VERSION=0
    PAT_VERSION=0
    ALPHA_DATE=$(date '+%m%d%y')
    ALPHA_VERSION=1
    ;;
  minor)
    MIN_VERSION=$((MIN_VERSION+1))
    PAT_VERSION=0
    ALPHA_DATE=$(date '+%m%d%y')
    ALPHA_VERSION=1
    ;;
  patch)
    PAT_VERSION=$((PAT_VERSION+1))
    ALPHA_DATE=$(date '+%m%d%y')
    ALPHA_VERSION=1
    ;;
  *)
    echo "Unknown bump version:  ${BUMP_VERSION}"
    echo ""
    bump_usage
    exit 6
    ;;
esac

BUMPED_VERSION="v${MAJ_VERSION}.${MIN_VERSION}.${PAT_VERSION}"
if [[ ${OMIT_ALPHA} -ne 1 ]]; then
    BUMPED_VERSION="${BUMPED_VERSION}-alpha-${ALPHA_DATE}.${ALPHA_VERSION}"
fi

if [ ${PROJECT_SILENT} -ne 1 ]; then
    echo -e "Bumped       : ${BUMPED_VERSION}"
fi

if [[ ${DRY_RUN} -eq 0 ]]; then
    git tag -a ${BUMPED_VERSION}
    git push --tags
else
    echo "${BUMPED_VERSION}"
    if [ ${PROJECT_SILENT} -ne 1 ]; then
        echo "git tag -a ${BUMPED_VERSION}"
        echo "git push --tags"
    fi
fi

