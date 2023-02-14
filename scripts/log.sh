#!/usr/bin/env bash

### Colour Variables

BLACK="30"
RED="31"
GREEN="32"
YELLOW="33"
BLUE="34"
MAGENTA="35"
CYAN="36"
L_GRAY="37"
GRAY="90"
L_RED="91"
L_GREEN="92"
L_YELLOW="93"
L_BLUE="94"

BG_BLACK="40"
BG_RED="41"
BG_GREEN="42"
BG_YELLOW="43"
BG_BLUE="44"
BG_MAGENTA="45"
BG_CYAN="46"
BG_L_GRAY="47"
BG_GRAY="100"
BG_L_RED="101"
BG_L_GREEN="102"
BG_L_YELLOW="103"
BG_L_BLUE="104"
L_MAGENTA="95"
BG_L_MAGENTA="105"
L_CYAN="96"
BG_L_CYAN="106"
WHITE="97"
BG_WHITE="107"

CLR_RESET="\033[0m"
BRIGHT="1"
FAINT="2"
ITALICS="3"
UNDER="4"
CLR_START="\033["
CLR_END="m"

create_info_log_line() {
  sep="===" 
  lnHdr="${CLR_START}${BRIGHT};${L_RED}${CLR_END}${sep}${CLR_RESET}"
  echo -e "${lnHdr}   $1   ${lnHdr}${CLR_RESET}"
}

log_info_stage() {
  stage="${CLR_START}${BRIGHT};${L_GRAY}${CLR_END}$1 project${CLR_RESET}"
  name="${CLR_START}${BRIGHT};${GREEN}${CLR_END}$2${CLR_RESET}"
  create_info_log_line "${stage} ${name}"
}

create_error_log_line() {
  sep="---"
  echo -e "${CLR_START}${BG_RED};${WHITE}${CLR_END}${sep}   $1    ${sep}${CLR_RESET}"
}

