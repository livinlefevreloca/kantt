#! /bin/bash
#
# This script is used to build the project.
#
# Usage:
#  ./build.sh [--no-cache]

cache=1
push=0
while getopts ":cpa:" opt; do
  case ${opt} in
    a )
      app=$OPTARG
      ;;
    c )
      echo "Building the project without cache."
      cache=0
      ;;
    p )
      echo "Pushing the images to the registry."
      push=1
      ;;
    \? )
      echo "Usage: cmd [-a]"
      exit 1
      ;;
  esac
done


if [[ $app == "kantt" || $app == "collector" ]]; then
  if [[ $cache == 0 ]]; then
    docker buildx build --no-cache --platform linux/amd64 --progress plain  -t kantt-collector .
  else
    docker buildx build --platform linux/amd64 --progress plain  -t kantt-collector .
  fi
  if [ "$?" -ne 0 ]; then
    echo "Failed to build the project."
    exit 1
  fi
fi


if [[ $app == "kantt" || $app == "dashboard" ]]; then
  if [[ $cache == 0 ]]; then
    docker buildx build  --no-cache --platform linux/amd64 --progress plain  -t kantt-dashboard -f ./Dockerfile.dashboard .
  else
    docker buildx build --platform linux/amd64 --progress plain  -t kantt-dashboard -f ./Dockerfile.dashboard .
  fi
  if [ "$?" -ne 0 ]; then
    echo "Failed to build the project."
    exit 1
  fi
fi

if [[ $push == 1 ]]; then
  docker tag kantt-collector livinlefevrel0ca/kantt-collector:latest
  docker tag kantt-dashboard livinlefevrel0ca/kantt-dashboard:latest
  docker push livinlefevrel0ca/kantt-collector:latest
  docker push livinlefevrel0ca/kantt-dashboard:latest
fi
