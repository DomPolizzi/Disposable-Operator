eval $(minikube docker-env)

docker build -t go-restart-operator:latest .

echo " Docker Image Built"