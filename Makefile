all: runshell

build:
	docker build -t dashboard-backend-img .

run: build
	mkdir -p data
	docker run -d --rm --name=dashboard-backend -p 8080:8080 -v $(PWD)/config.yaml:/home/config.yaml:ro -v $(PWD)/data:/home/data dashboard-backend-img

runshell: build
	mkdir -p data
	docker run -it --rm --name=dashboard-backend -p 8080:8080 -v $(PWD)/config.yaml:/home/config.yaml:ro -v $(PWD)/data:/home/data dashboard-backend-img /bin/sh
