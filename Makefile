MAKEFLAGS += --warn-undefined-variables
SHELL     := /bin/bash -e -u -o pipefail
SVC       := github.com
ORG       := supercaracal
REPO      := kubernetes-controller-template
MOD_PATH  := ${SVC}/${ORG}/${REPO}
IMG_TAG   := latest
REGISTRY  := 127.0.0.1:5000
TEMP_DIR  := _tmp

all: build test lint

${TEMP_DIR}:
	@mkdir -p $@

${TEMP_DIR}/codegen: CURRENT_DIR           := $(shell pwd)
${TEMP_DIR}/codegen: GOBIN                 ?= $(shell go env GOPATH)/bin
${TEMP_DIR}/codegen: GOENV                 += GOROOT=${CURRENT_DIR}/${TEMP_DIR}
${TEMP_DIR}/codegen: LOG_LEVEL             ?= 1
${TEMP_DIR}/codegen: API_VERSION           := v1
${TEMP_DIR}/codegen: CODE_GEN_INPUT        := ${MOD_PATH}/pkg/apis/${ORG}/${API_VERSION}
${TEMP_DIR}/codegen: CODE_GEN_OUTPUT       := ${MOD_PATH}/pkg/generated
${TEMP_DIR}/codegen: CODE_GEN_ARGS         += --output-base=${CURRENT_DIR}/${TEMP_DIR}/src
${TEMP_DIR}/codegen: CODE_GEN_ARGS         += --go-header-file=${CURRENT_DIR}/${TEMP_DIR}/empty.txt
${TEMP_DIR}/codegen: CODE_GEN_ARGS         += -v ${LOG_LEVEL}
${TEMP_DIR}/codegen: CODE_GEN_DEEPC        := zz_generated.deepcopy
${TEMP_DIR}/codegen: CODE_GEN_CLI_SET_NAME := versioned
${TEMP_DIR}/codegen: ${TEMP_DIR} $(shell find pkg/apis/${ORG}/ -type f -name '*.go')
	@touch -a ${TEMP_DIR}/empty.txt
	@mkdir -p ${TEMP_DIR}/src/${MOD_PATH}
	@ln -sf ${CURRENT_DIR}/pkg ${TEMP_DIR}/src/${MOD_PATH}/
	@# https://github.com/kubernetes/gengo/blob/master/args/args.go
	@# https://github.com/kubernetes/code-generator/tree/master/cmd
	${GOENV} ${GOBIN}/deepcopy-gen ${CODE_GEN_ARGS} --input-dirs=${CODE_GEN_INPUT} --bounding-dirs=${CODE_GEN_INPUT} --output-file-base=${CODE_GEN_DEEPC}
	${GOENV} ${GOBIN}/client-gen   ${CODE_GEN_ARGS} --input=${CODE_GEN_INPUT}      --output-package=${CODE_GEN_OUTPUT}/clientset --input-base="" --clientset-name=${CODE_GEN_CLI_SET_NAME}
	${GOENV} ${GOBIN}/lister-gen   ${CODE_GEN_ARGS} --input-dirs=${CODE_GEN_INPUT} --output-package=${CODE_GEN_OUTPUT}/listers
	${GOENV} ${GOBIN}/informer-gen ${CODE_GEN_ARGS} --input-dirs=${CODE_GEN_INPUT} --output-package=${CODE_GEN_OUTPUT}/informers --versioned-clientset-package=${CODE_GEN_OUTPUT}/clientset/${CODE_GEN_CLI_SET_NAME} --listers-package=${CODE_GEN_OUTPUT}/listers
	@touch $@

codegen: ${TEMP_DIR}/codegen

build: GOOS        ?= $(shell go env GOOS)
build: GOARCH      ?= $(shell go env GOARCH)
build: CGO_ENABLED ?= $(shell go env CGO_ENABLED)
build: FLAGS       += -ldflags="-s -w"
build: FLAGS       += -trimpath
build: FLAGS       += -tags timetzdata
build: codegen
	GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=${CGO_ENABLED} go build ${FLAGS} -o ${REPO}

test:
	@go clean -testcache
	@go test -race ./...

lint:
	@go vet ./...
	@golint -set_exit_status ./...

run: TZ  ?= Asia/Tokyo
run: CFG ?= $$HOME/.kube/config
run:
	@TZ=${TZ} ./${REPO} --kubeconfig=${CFG}

clean:
	@unlink ${TEMP_DIR}/src/${MOD_PATH}/pkg || true
	@rm -rf ${REPO} main ${TEMP_DIR}

build-image:
	@docker build -t ${REPO}:${IMG_TAG} .

lint-image:
	@docker run --rm -i hadolint/hadolint < Dockerfile

port-forward:
	@kubectl --context=kind-kind port-forward service/registry 5000:5000

push-image:
	@docker tag ${REPO}:${IMG_TAG} ${REGISTRY}/${REPO}:${IMG_TAG}
	@docker push ${REGISTRY}/${REPO}:${IMG_TAG}

clean-image:
	@docker rmi -f ${REPO}:${IMG_TAG} ${REGISTRY}/${REPO}:${IMG_TAG} || true
	@docker image prune -f
	@docker volume prune -f

apply-manifests:
	@kubectl --context=kind-kind apply -f config/registry.yaml
	@kubectl --context=kind-kind apply -f config/crd.yaml
	@kubectl --context=kind-kind apply -f config/example-foobar.yaml
	@kubectl --context=kind-kind apply -f config/controller.yaml

replace-k8s-go-module: KUBE_LIB_VER := 1.22.1
replace-k8s-go-module:
	@./scripts/replace_k8s_go_module.sh ${KUBE_LIB_VER}

wait-registry-running:
	@./scripts/wait_pod_running.sh registry

wait-controller-running:
	@./scripts/wait_pod_running.sh controller
