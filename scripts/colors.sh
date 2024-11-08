#!/usr/bin/env bash

# Intended use is to 'source' this file and use the functions provided.

#
# Functions:
#

# color [-bg-<color>] [-<color>] [-<effect> [-<effect>]...] text text text
# where
#   -bg-<color>:    applies the named color as the text background
#                   (only applies the last one specified)
#   -<color>:       applies the named color as the foreground
#   -<effect>:      applies the named effect(s) to the text
#   text text text: text to colorize
#
# Examples:
#   echo "$(color -italics -green text)"
#   => prints "text" in green italics
#
#   echo "$(color -bright -red ERROR:) this is bad news"
#   => prints "ERROR" in bright red and " this is bad news" in the default colour
#
#   echo "$(color -blue -underline shazam)"
#   => prints "shazam" in blue with an underline
#
#   echo "$(color -bg-red -bright -blue canberra)"
#   => prints "canberra" in bright blue on a red background
#
#   echo "$(color -bg-blue -bg-red hello)"
#   => prints "hello" on a red backbround
#
#   echo "$(color -bg blue -bright -yellow hello)"
#   => prints "hello" in bright yellow on a blue background
#
#   echo "$(color -bg blue -bright -yellow -underline hello)"
#   => prints "hello" in underlined bright yellow on a blue background
#
color() {
  # Control
  local _c_start="\033["
  local _c_end="m"
  local _c_reset="${_c_start}0${_c_end}"

  # Foregrounds
  # dark
  local _fg_black="30"
  local _fg_red="31"
  local _fg_green="32"
  local _fg_yellow="33"
  local _fg_blue="34"
  local _fg_magenta="35"
  local _fg_cyan="36"
  local _fg_gray="90"
  local _fg_grey="90"
  # light
  local _fg_lt_gray="37"
  local _fg_lt_grey="37"
  local _fg_lt_red="91"
  local _fg_lt_green="92"
  local _fg_lt_yellow="93"
  local _fg_lt_blue="94"
  local _fg_lt_magenta="95"
  local _fg_lt_cyan="96"
  local _fg_white="97"

  # Backgrounds
  # dark
  local _bg_black="40"
  local _bg_red="41"
  local _bg_green="42"
  local _bg_yellow="43"
  local _bg_blue="44"
  local _bg_magenta="45"
  local _bg_cyan="46"
  local _bg_gray="100"
  local _bg_grey="100"
  # light
  local _bg_lt_gray="47"
  local _bg_lt_grey="47"
  local _bg_lt_red="101"
  local _bg_lt_green="102"
  local _bg_lt_yellow="103"
  local _bg_lt_blue="104"
  local _bg_lt_magenta="105"
  local _bg_lt_cyan="106"
  local _bg_white="107"

  # effects (note: we support some 'similar' names for some values)
  # actual support may depend on your terminal
  local _e_bright="1"
  local _e_bold="1"
  local _e_faint="2"
  local _e_dark="2"
  local _e_italics="3"
  local _e_italic="3"
  local _e_underscore="4"
  local _e_underline="4"
  local _e_blink="5" # slow blink, my not be supported by the terminal
  local _e_flash="6" # fast blink, my not be supported by the terminal
  local _e_inverse="7"
  local _e_hide="8" # leaves blanks for the string length, highlight does not show the value
  local _e_hidden="8" # leaves blanks for the string length, highlight does not show the value
  local _e_strikeout="9"
  local _e_strike="9"

  #
  # start logic
  #

  local _text=
  local _color_codes=

  while [ $# -gt 0 ]; do
    case $1 in
      -bg-*) # "*" must be a color name
        local _value=${1:4}
        _cname="_bg_${_value}"
        [ -n "${!_cname}" ] && _color_codes+="${!_cname} "
        ;;
      -*) # "*" may be a name (color or effect), or it's just more text
        local _value=${1:1}
        local _cname="_fg_${_value}"
        local _ename="_e_${_value}"
        if [ -n "${!_cname}" ]; then
          _color_codes+="${!_cname} "
        elif [ -n "${!_ename}" ]; then
          _color_codes+="${!_ename} "
        else
          # no match: assume this is raw text
          [ -n "${_text}" ] && _text+=" "
          _text+="${1}"
        fi
        ;;
      *)
        [ -n "${_text}" ] && _text+=" "
        _text+="${1}"
        ;;
    esac
    shift
  done

  # Allows effect, bg color and fg color to be specified in any order
  #   e.g. $(color -bright -red foo) and $(color -red -bright foo) are the same
  local _color=
  for c in ${_color_codes}; do
    _color+="$c;"
  done
  local _msg="${_c_start}${_color%%;}${_c_end}${_text}${_c_reset}"
  # echo "debug: msg = '${_msg}'"
  echo -e "${_msg}"
}

# Print "WARNING: $*" where 'WARNING' is in yellow
warning() {
  echo "$(color -bold -yellow WARNING): $*"
}

# Print "ERROR: $*", where 'ERROR' is red
error() {
  echo "$(color -bold -lt_red ERROR): $*"
}