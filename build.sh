#! /bin/bash
#
# This script is used to build the project.
#
# Usage:
#  ./build.sh [--no-cache]

if [[ "$1" == '--no-cache' ]]; then
docker buildx build --no-cache --platform linux/amd64 --progress plain  -t kantt-collector .
else
docker buildx build --platform linux/amd64 --progress plain  -t kantt-collector .
fi
if [ "$?" -ne 0 ]; then
  echo "Failed to build the project."
  exit 1
fi
docker tag kantt-collector livinlefevrel0ca/kantt-collector:latest
docker push livinlefevrel0ca/kantt-collector:latest
