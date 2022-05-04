-include ../../version
-include ../Makefile.options
#####################################################################################
service=$(docker_prefix)upload-service
version=$(upload_version)
dir=upload
########### DOCKER ##################################################################
tag=$(service):$(version)

dbuild: $(dist_dir)/$(executable_name)
	cd ../../ && docker build -t $(tag) --build-arg BUILD_VERSION=$(version) -f build/$(dir)/Dockerfile .

dpush: dbuild
	docker push $(tag)

dscan: dbuild
	docker scan --accept-license $(tag)		
#####################################################################################
.PHONY:
	dbuild dpush dscan
