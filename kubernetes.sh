export DOCKER_IMAGE=stevelacy/kubermaster:latest
export APP_NAME=kubermaster
export TOKEN=test-token
export BUILD_TAG=$(date)

# this ensures that the context uses the locally built docker image
eval $(minikube docker-env)

# build the image in the minikube env
./build.sh

sigil -p -f kubermaster.yaml | kubectl apply -f -
