#!/bin/bash

#ws docker registry push
export CTR_REGISTRY=docker.dev.ws:5000
export CTR_TAG=osmv0.9.1-tag6

make docker-push-init-osm-controller
make docker-push-osm-controller
make docker-push-init
make docker-push-osm-injector

#aws docker push
export CTR_REGISTRY=978944737929.dkr.ecr.us-west-2.amazonaws.com
make docker-push-init-osm-controller
make docker-push-osm-controller
make docker-push-init
make docker-push-osm-injector
