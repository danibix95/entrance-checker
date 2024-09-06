docker_bin := docker
docker_build_bin := $(docker_bin) buildx build

dockerfile_filepath := .docker/Dockerfile

docker_compose := $(docker_bin) compose
docker_compose_file := .docker/docker-compose.yml
docker_compose_flags := -f $(docker_compose_file)


up: # start the selected docker-compose service. If parameter c is not specified, stop and remove all services
	@$(docker_compose) $(docker_compose_flags) $(env) up --remove-orphans -d $(c)
.PHONY: up

down: # stop and remove the selected docker-compose service. If parameter c is not specified, stop and remove all services
	@$(docker_compose) $(docker_compose_flags) down --remove-orphans $(c)
.PHONY: down

