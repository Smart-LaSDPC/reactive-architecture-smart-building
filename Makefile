# DOCKER_REGISTRY=andromeda.lasdpc.icmc.usp.br:50010
DOCKER_REPO=ppastorf/tcc-etl
INGESTOR_TAG=ingestor
PUBLISHER_TAG=publisher

local-up:
	@docker compose -f docker-compose.yaml up -d --build
local-down:
	@docker compose -f docker-compose.yaml down
local-logs:
	@docker compose logs --follow

build-ingestor:
	@docker build ./go-ingestor -t $(INGESTOR_TAG) -f ./go-ingestor/Dockerfile
push-ingestor:
	@docker tag $(INGESTOR_TAG) $(DOCKER_REPO):$(INGESTOR_TAG)
	@docker push $(DOCKER_REPO):$(INGESTOR_TAG)
build-publisher:
	@docker build ./mosquitto-publisher -t $(PUBLISHER_TAG) -f ./mosquitto-publisher/Dockerfile
push-publisher:
	@docker tag $(PUBLISHER_TAG) $(DOCKER_REPO):$(PUBLISHER_TAG)
	@docker push $(DOCKER_REPO):$(PUBLISHER_TAG)

publish-k8s:
	@docker compose -f compose-publisher-k8s.yaml up --build
