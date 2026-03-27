#!/usr/bin/env bash

DIRECTORY="$(cd "$(dirname "${0}" > /dev/null)" && pwd || exit 1)"
CLUSTER_NAME="yeetcd"

__check_dependency() {
  if ! which "${1}" >/dev/null; then
    echo "Please install ${1}"
    exit 1
  fi
}

__check_dependencies() {
  __check_dependency k3d
  __check_dependency helm
}

__create_cluster() {
  if ! k3d cluster list | grep -q "${CLUSTER_NAME}"; then
    VOLUME_ARG=""
    if [[ -f "${DIRECTORY}/trusted-cas.pem" ]]; then
      VOLUME_ARG="--volume '${DIRECTORY}/trusted-cas.pem:/etc/ssl/certs/additional-ca.crt'"
    fi
    eval "k3d cluster create '${CLUSTER_NAME}' --registry-create ${CLUSTER_NAME}-registry --agents 1 ${VOLUME_ARG}"
  fi
}

__registry_port() {
  docker inspect "${CLUSTER_NAME}-registry" | jq -r '.[0].NetworkSettings.Ports["5000/tcp"][0].HostPort'
}

__delete_cluster() {
  k3d cluster delete "${CLUSTER_NAME}" || true
}

__helm_upgrade() {
  helm upgrade --install --reset-values --set 'local=true' yeetcd ./helm
}

start() {
  __check_dependencies
  __create_cluster
  __helm_upgrade
  __create_test_config
}

__create_test_config() {
  mkdir -p "${DIRECTORY}/controller/src/test/resources"
  cat <<yeetcd-controller > "${DIRECTORY}/controller/src/test/resources/yeetcd-controller.yaml"
registry:
  pushAddress: localhost:$(__registry_port)
  pullAddress: ${CLUSTER_NAME}-registry:5000
yeetcd-controller
  k3d kubeconfig get "${CLUSTER_NAME}" > "${DIRECTORY}/controller/src/test/resources/kubeconfig"
}

__remove_test_config() {
  rm "${DIRECTORY}/controller/src/test/resources/yeetcd-controller.yaml" || true
  rm "${DIRECTORY}/controller/src/test/resources/kubeconfig" || true
}

update() {
  __helm_upgrade
}

stop() {
  __delete_cluster
  __remove_test_config
}

__list_all_commands() {
  for func in $(compgen -A function | grep -v "__"); do
    echo "$func"
  done
}

__usage() {
  echo -e "One of the following commands is required:"
  __list_all_commands
  exit 1
}

if [[ $# -ne 1 ]]; then
  __usage
elif __list_all_commands | grep -q "^${1}$"; then
  if ! cd "${DIRECTORY}"; then
    echo "Unable to cd to ${DIRECTORY}"
    exit 1
  fi
  eval "${1}"
else
  __usage
fi
