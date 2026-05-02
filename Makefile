.PHONY: proto-gen proto-lint proto-breaking proto-format deps-update

PROTO_DIR := proto

proto-gen:
	cd $(PROTO_DIR) && buf generate

proto-lint:
	cd $(PROTO_DIR) && buf lint

proto-format:
	cd $(PROTO_DIR) && buf format -w

proto-breaking:
	cd $(PROTO_DIR) && buf breaking --against '../.git#branch=main,subdir=proto'

deps-update:
	cd $(PROTO_DIR) && buf dep update

proto-all: proto-lint proto-gen