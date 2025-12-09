all: runshell

build:
	docker build -t dashboard-backend-img .

run: build
	docker run -d --rm --name=dashboard-backend -p 8080:8080 dashboard-backend-img

runshell: build
	docker run -it --rm --name=dashboard-backend -p 8080:8080 dashboard-backend-img /bin/sh
