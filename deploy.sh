if [[ "$1" != "" ]]; then
    BUILD_ID="$1"
else
    echo ERROR: Invalid BUILD_ID argument: ["$1"] 1>&2
    exit 1 # terminate and indicate error
fi

docker ps -f name=dashboard-backend    
docker pull kovadocker/dashboard-backend:${BUILD_ID}
docker stop dashboard-backend
docker rm dashboard-backend
docker run --name=dashboard-backend --restart=always -p 8888:8080 -d kovadocker/dashboard-backend:${BUILD_ID}
