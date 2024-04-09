REGISTRY := ""

define docker_build
    if [ -n "$(REGISTRY)" ]; then \
        docker build $(1)/ -f Dockerfile \
            --build-arg APP=$(1).py \
            -t $(REGISTRY)/$(1); \
    else \
        docker build $(1)/ -f Dockerfile \
            --build-arg APP=$(1).py \
            -t $(1); \
    fi
endef

define docker_push
    if [ -n "$(REGISTRY)" ]; then \
        docker push $(REGISTRY)/$(1); \
    else \
        docker push $(1); \
    fi
endef

define kind_load
    if [ -n "$(REGISTRY)" ]; then \
        kind load docker-image $(REGISTRY)/$(1); \
    else \
        kind load docker-image $(1); \
    fi
endef

docker-build:
	$(call docker_build,prepare-data)
	$(call docker_build,preprocess-data)
	$(call docker_build,train-data)
	$(call docker_build,test-data)

push-images:
	$(call docker_push,prepare-data)
	$(call docker_push,preprocess-data)
	$(call docker_push,train-data)
	$(call docker_push,test-data)

load-images:
	$(call kind_load,prepare-data)
	$(call kind_load,preprocess-data)
	$(call kind_load,train-data)
	$(call kind_load,test-data)
