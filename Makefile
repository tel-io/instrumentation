include ./Makefile.Common

# ALL_MODULES includes ./* dirs (excludes . dir)
ALL_MODULES := $(shell find . -type f -name "go.mod" -exec dirname {} \; | sort | egrep  '^./' )

.PHONY: all
all: gotest gofmt update-tel gotidy

.PHONY: gomoddownload
gomoddownload:
	@$(MAKE) for-all-target TARGET="moddownload"

.PHONY: gotest
gotest:
	@$(MAKE) for-all-target TARGET="test test-unstable"

.PHONY: gotest-with-cover
gotest-with-cover:
	@$(MAKE) for-all-target TARGET="test-with-cover"
	$(GOCOVMERGE) $$(find . -name coverage.out) > coverage.txt

.PHONY: golint
golint:
	@$(MAKE) for-all-target TARGET="lint lint-unstable"

.PHONY: goimpi
goimpi:
	@$(MAKE) for-all-target TARGET="impi"

.PHONY: gofmt
gofmt:
	@$(MAKE) for-all-target TARGET="fmt"

.PHONY: gotidy
gotidy:
	@$(MAKE) for-all-target TARGET="tidy"

.PHONY: update-tel
update-tel:
	@$(MAKE) for-all-target TARGET="updatetel"

all-modules:
	@echo $(ALL_MODULES) | tr ' ' '\n' | sort

# Append root module to all modules
GOMODULES = $(ALL_MODULES)

# Define a delegation target for each module
.PHONY: $(GOMODULES)
$(GOMODULES):
	@echo "Running target '$(TARGET)' in module '$@'"
	$(MAKE) -C $@ $(TARGET)

# Triggers each module's delegation target
.PHONY: for-all-target
for-all-target: $(GOMODULES)