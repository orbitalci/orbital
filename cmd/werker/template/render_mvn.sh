#!/bin/bash

# order of arguments: MAVEN_SETTINGS_XML
if [ $# -gt 0 ]; then
  args=("$@")
  mvnsettings=${args[0]}
  if [ ! -z "${mvnsettings}" ]; then
    mkdir -p ~/.m2/
    echo ${mvnsettings} >> ~/.m2/settings.xml
  else
    echo "maven settings variable empty, saving nothing to settings.xml"
  fi
else
    echo "no arguments were passed in"
fi