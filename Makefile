NAME 			:= persistence
BIN         	:= $(NAME)
REPO        	:= farm.e-pedion.com/repo/$(NAME)
BUILD       	:= $(shell openssl rand -hex 10)
VERSION     	:= $(shell if [ -f version ]; then awk '{printf $1}' < version; else openssl rand -hex 5; fi)
MAKEFILE    	:= $(word $(words $(MAKEFILE_LIST)), $(MAKEFILE_LIST))
BASE_DIR    	:= $(shell cd $(dir $(MAKEFILE)); pwd)
ALLPKGS        	:= $(shell go list ./... | grep -v /vendor/)
PKGS        	:= $(shell go list ./... | grep -v /vendor/ | grep -v /itest)
IPKGS        	:= $(shell go list ./... | grep -v /vendor/ | grep /itest)

#Test and Benchmark Parameters
TEST_PKGS ?= 
COVERAGE_FILE := $(NAME).coverage
COVERAGE_HTML := $(NAME).coverage.html
PKG_COVERAGE := $(NAME).pkg.coverage

.PHONY: default
default: build

.PHONY: install
install: install_sw_deps sync_deps
	@echo "$(REPO) installed successfully" 

.PHONY: install_sw_deps
install_sw_deps:
	brew install go
	go get -u github.com/kardianos/govendor

.PHONY: install_deps
install_deps:
	govendor fetch github.com/matryer/resync
	govendor fetch github.com/Sirupsen/logrus/...
	govendor fetch github.com/uber-go/zap

.PHONY: sync_deps
sync_deps:
	govendor sync

.PHONY: all
all: build test bench_all coverage

.PHONY: build
build:
	@echo "Building $(REPO)@$(VERSION)-$(BUILD)"
	go build $(PKGS)

.PHONY: clean
clean: 
	-rm $(NAME)*coverage*
	-rm *.test
	-rm *.pprof

.PHONY: reset
reset: clean 
	-cd vendor; rm -r */

.PHONY: test_all
test_all:
	go test -v -race $(ALLPKGS)

.PHONY: test
test:
	@if [ "$(TEST_PKGS)" == "" ]; then \
	    echo "Unit Test All Pkgs";\
		go test -v -race $(PKGS) || exit 501;\
	else \
	    echo "Unit Test Selected Pkgs=$(TEST_PKGS)";\
	    for tstpkg in $(TEST_PKGS); do \
			go test -v -race $(REPO)/$$tstpkg || exit 501;\
		done; \
	fi

.PHONY: itest
itest:
	@if [ "$(TEST_PKGS)" == "" ]; then \
	    echo "Integration Test All Pkgs";\
		go test -v -race $(IPKGS) || exit 501;\
	else \
	    echo "Integration Unit Test Selected Pkgs=$(TEST_PKGS)";\
	    for tstpkg in $(TEST_PKGS); do \
			go test -v -race $(REPO)/$$tstpkg/itest || exit 501;\
		done; \
	fi

.PHONY: bench_all
bench_all:
	#go test -bench=. -run="^$$" -cpuprofile=cpu.pprof -memprofile=mem.pprof -benchmem $(IPKGS)
	go test -bench=. -run="^$$" -benchmem $(PKGS)

.PHONY: benchmark
bench:
	@if [ "$(TEST_PKGS)" == "" ]; then \
	    echo "Benchmark all Pkgs" ;\
	    for tstpkg in $(IPKGS); do \
		    go test -bench=. -run="^$$" -cpuprofile=cpu.pprof -memprofile=mem.pprof -benchmem $$tstpkg || exit 501;\
		done; \
	else \
	    echo "Benchmark Selected Pkgs=$(TEST_PKGS)" ;\
	    for tstpkg in $(TEST_PKGS); do \
		    go test -bench=. -run="^$$" -cpuprofile=cpu.pprof -memprofile=mem.pprof -benchmem $(REPO)/$$tstpkg/itest || exit 501;\
		done; \
	fi

.PHONY: coverage
coverage:
	@echo "Testing with coverage"
	@echo 'mode: set' > $(COVERAGE_FILE)
	@touch $(PKG_COVERAGE)
	@touch $(COVERAGE_FILE)
	@if [ "$(TEST_PKGS)" == "" ]; then \
		for pkg in $(ALLPKGS); do \
			go test -v -coverprofile=$(PKG_COVERAGE) $$pkg || exit 501; \
			if (( `grep -c -v 'mode: set' $(PKG_COVERAGE)` > 0 )); then \
				grep -v 'mode: set' $(PKG_COVERAGE) >> $(COVERAGE_FILE); \
			fi; \
		done; \
	else \
	    echo "Testing with covegare the Pkgs=$(TEST_PKGS)" ;\
	    for tstpkg in $(TEST_PKGS); do \
			go test -v -coverprofile=$(PKG_COVERAGE) $(REPO)/$$tstpkg || exit 501; \
			if (( `grep -c -v 'mode: set' $(PKG_COVERAGE)` > 0 )); then \
				grep -v 'mode: set' $(PKG_COVERAGE) >> $(COVERAGE_FILE); \
			fi; \
		done; \
	fi
	@echo "Generating report"
	@go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	open $(COVERAGE_HTML)
