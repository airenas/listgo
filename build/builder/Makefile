-include ../../version
#####################################################################################
service=airenas/list-builder
version=$(builder_version)
########### DOCKER ##################################################################
tag=$(service):$(version)

dbuild: $(dist_dir)/$(executable_name)
	cd ../../ && docker build -t $(tag) --build-arg BUILD_VERSION=$(version) -f build/builder/Dockerfile .

dpush: dbuild
	docker push $(tag)

dscan: dbuild
	docker scan --accept-license $(tag)		
#####################################################################################
.PHONY:
	dbuild dpush dscan
