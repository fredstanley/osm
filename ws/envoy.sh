#export CTR_REGISTRY=docker.dev.ws:5000
#aws docker push
#export CTR_REGISTRY=978944737929.dkr.ecr.us-west-2.amazonaws.com
#docker registry
docker pull envoyproxy/envoy-alpine:v1.18.4
docker tag envoyproxy/envoy-alpine:v1.18.4 docker.dev.ws:5000/envoy-alpine:v1.18.4
docker push docker.dev.ws:5000/envoy-alpine:v1.18.4

#aws
docker pull envoyproxy/envoy-alpine:v1.18.4
docker tag envoyproxy/envoy-alpine:v1.18.4 978944737929.dkr.ecr.us-west-2.amazonaws.com/envoy-alpine:v1.18.4
docker push 978944737929.dkr.ecr.us-west-2.amazonaws.com/envoy-alpine:v1.18.4
