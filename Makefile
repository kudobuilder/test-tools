.PHONY: test
test: generate
	@./scripts/test.sh

.PHONY: generate
generate:
	@./scripts/generate.sh

.PHONY: lint
lint: generate
	@./scripts/lint.sh

.PHONY: clean
clean:
	@rm -rf bin
