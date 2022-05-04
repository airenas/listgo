-include ../../version
-include ../Makefile.options
#####################################################################################
service=$(docker_prefix)manager-service
version=$(manager_version)
dir=manager
########### DOCKER ##################################################################
tag=$(service):$(version)

dbuild:
	cd ../../ && docker build -t $(tag) --build-arg BUILD_VERSION=$(version) -f build/$(dir)/Dockerfile .

dpush: dbuild
	docker push $(tag)

dscan: dbuild
	docker scan --accept-license $(tag)		
#####################################################################################
.PHONY:
	dbuild dpush dscan
