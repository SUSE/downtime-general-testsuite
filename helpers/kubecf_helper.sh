#!/bin/bash

if [ -z "$1" ]; then
  echo "Usage: $0 <deploy|upgrade|clean>"
  echo
  echo " deploy   - install kubecf with given chart from KUBECF_CHART"
  echo " upgrade  - upgrade existing kubecf with given chart from KUBECF_CHART"
  echo " password - get the admin password"
  echo " clean    - uninstall kubecf"
  exit 1
fi
set -e -x

CMD="$1"

if [[ -z "${KUBECF_CHART}" && "$CMD" != "clean" ]]; then
  echo "\$KUBECF_CHART is not set. Please set to the path of the kubecf chart."; exit 1
fi

if [ -z "${CATAPULT_DIR}" ]; then
  echo "\$CATAPULT_DIR is not set. Please set to the path of the cloned catapult repository."; exit 1
fi

pushd "${CATAPULT_DIR}"

case "${CMD}" in
  deploy)
    SCF_CHART="${KUBECF_CHART}" make kubecf
    ;;
  upgrade)
    SCF_CHART="${KUBECF_CHART}" make kubecf-upgrade
    ;;
  clean)
    make kubecf-clean
    ;;
esac
