if [[ "$1" != "" ]]; then
    BUILD_ID="$1"
else
    echo ERROR: Invalid BUILD_ID argument: ["$1"] 1>&2
    exit 1 # terminate and indicate error
fi

docker network inspect mynet >/dev/null 2>&1 || docker network create --driver bridge mynet

docker ps -f name=dashboard-backend    
docker pull kovadocker/dashboard-backend:${BUILD_ID}
docker stop dashboard-backend || true
docker rm dashboard-backend || true
docker run --name=dashboard-backend --restart=always --network=mynet -p 8888:8080 -d kovadocker/dashboard-backend:${BUILD_ID}
