eval $(minikube docker-env)

docker stop go-restart-operator
docker rm go-restart-operator

docker rmi go-restart-operator