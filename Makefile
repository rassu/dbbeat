BEAT_NAME=dbbeat
BEAT_PATH=github.com/rassu/dbbeat
BEAT_GOPATH=$(firstword $(subst :, ,${GOPATH}))
SYSTEM_TESTS=false
TEST_ENVIRONMENT=false
ES_BEATS_IMPORT_PATH=github.com/elastic/beats/v7
ES_BEATS?=$(shell go list -m -f '{{.Dir}}' ${ES_BEATS_IMPORT_PATH})
LIBBEAT_MAKEFILE=$(ES_BEATS)/libbeat/scripts/Makefile
GOPACKAGES=$(shell go list ${BEAT_PATH}/... | grep -v /tools)
GOBUILD_FLAGS=-i -ldflags "-X ${ES_BEATS_IMPORT_PATH}/libbeat/version.buildTime=$(NOW) -X ${ES_BEATS_IMPORT_PATH}/libbeat/version.commit=$(COMMIT_ID)"
MAGE_IMPORT_PATH=github.com/magefile/mage
NO_COLLECT=true
CHECK_HEADERS_DISABLED=true

# Path to the libbeat Makefile
-include $(LIBBEAT_MAKEFILE)

.PHONY: copy-vendor
copy-vendor:
	mage vendorUpdate


.PHONY: install-mage
install-mage:
	cd ./mage &&  go run bootstrap.go

.PHONY: docker
docker:
	docker run --rm -d --name some-postgres -e POSTGRES_PASSWORD=pwd -p 5432:5432 postgres && (cd docker-elk && docker-compose up -d kibana elasticsearch)

.PHONY: stop-docker
stop-docker:
	docker stop some-postgres && (cd docker-elk && docker-compose stop kibana elasticsearch)

.PHONY: test-create-db
test-create-db:
	cd examples && BEAT_STRICT_PERMS=false go run main.go create

.PHONY: test-insert-db
test-insert-db:
	cd examples && BEAT_STRICT_PERMS=false go run main.go insert



build:
	go build -o dbbeat main.go

run:
	BEAT_STRICT_PERMS=false go run main.go  -e -d "*"
