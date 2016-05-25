#!/bin/bash -e

app_name="azure_arm_proxy"
gitref=`git rev-parse --verify HEAD`

# Perform a continuous integration build. This involves the following steps:
#   - decide whether we need an image based on $1 (tag name) and $2 (pull request number)
#   - download bare Docker binary into bin; add bin to path
#   - build and push an image
ci ()
{
  if [[ "$2" == "false" ]]
  then
    pull $1 2>/dev/null || pull latest 2>/dev/null || :
    build $1
    push $1
  else
    echo "Skipping Docker image build due to pull-request status ($2)"
  fi
}
# Build a docker image, including performing any prerequisite work for the build: gathering
# dependencies, creating intermediate build artifacts in a build container, etc.
build ()
{
  echo "Building Docker image rightscale/$app_name:$1"
  echo "docker build --build-arg gitref=$gitref --tag rightscale/$app_name:$1"
  docker build --build-arg gitref=$gitref --tag rightscale/$app_name:$1 .
  return $?
}
# Login to DockerHub
login ()
{
  if [ -z "${DOCKERHUB_USER}${DOCKERHUB_EMAIL}${DOCKERHUB_PASSWORD}" ]
  then
    echo "ERROR: must export DOCKERHUB_(EMAIL|PASSWORD|USER) to perform this action"
    exit 40
  else
    echo "Logging into DockerHub as $DOCKERHUB_USER"
    docker login -e $DOCKERHUB_EMAIL -u $DOCKERHUB_USER -p $DOCKERHUB_PASSWORD
  fi
}
# Clean up intermediate build artifacts
clean ()
{
  echo "Removing intermediate build artifacts"
  rm -Rf build/*
}
# Pull a current tag of this repo. Can be used as layer cache for building an image.
pull ()
{
  login

  echo "Pulling docker image rightscale/$app_name:$1"
  docker pull rightscale/$app_name:$1

  return $?
}
# Push a named tag of this repo's image to DockerHub. This is a nearly-useless shortcut.
push ()
{
  login

  echo "Pushing Docker image rightscale/$app_name:$1"
  docker push rightscale/$app_name:$1
  return $?
}
help ()
{
  echo
  echo "Usage:"
  echo " $0 build [tag]"
  echo " $0 clean"
  echo " $0 push [tag]"
  echo " $0 pull [tag]"
  echo " $0 ci <branch> <pull_request_number|false>"
  echo
  echo "Default tag name is current git branch ($1)"
}
# Figure out which tag to use for this image build; default to git branch name if not specified
if [ -z "$2" ]
then
  git_ref=$(git symbolic-ref HEAD 2>/dev/null)
  git_branch=${git_ref##refs/heads/}
  tag=$git_branch
else
  tag=$2
fi
# Special case: 'master' branch (or tag) always corresponds to 'latest' tag
if [ "$tag" == "master" ]
then
  tag=latest
fi
pull_request_number=$3
# Run the command
case $1 in
ci)
  ci $tag $pull_request_number
  ;;
build)
  build $tag
  ;;
clean)
  clean $tag
  ;;
push)
  push $tag
  ;;
pull)
  pull $tag
  ;;
*)
  help $tag
  ;;
esac
