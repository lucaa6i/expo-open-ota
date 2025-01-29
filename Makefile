DOCKER_FLAG := $(findstring docker, $(MAKECMDGOALS))
HTML_FLAG := $(findstring html, $(MAKECMDGOALS))
MAKEFLAGS += --silent

build:
ifeq ($(DOCKER_FLAG),docker)
	docker-compose build
else
	go build ./...
endif

up:
ifeq ($(DOCKER_FLAG),docker)
	docker-compose up -d
else
	reflex -r '\.go$$' -s -- sh -c "go run cmd/api/main.go"
endif

down:
ifeq ($(DOCKER_FLAG),docker)
	docker-compose down
else
	echo "Not applicable locally. Stop the application manually."
endif

test_app:
ifeq ($(DOCKER_FLAG),docker)
	docker-compose exec ota-server sh -c "$(MAKE_COVERAGE_CMD)"
else
	$(MAKE_COVERAGE_CMD)
endif

test_app_watch:
	find . -name '*.go' | entr -n -c $(MAKE) test_app $(DOCKER_FLAG) $(HTML_FLAG)


define MAKE_COVERAGE_CMD
	go test -v -coverprofile=coverage.out ./... && \
	$(call CLEAN_COVERAGE) && \
	$(call GENERATE_HTML)
endef

define CLEAN_COVERAGE
	if [ "$(shell uname -s)" = "Darwin" ]; then \
		sed -i '' -e '/test/d' -e '/commands/d' -e '/cmd/d' -e '/lambda/d' coverage.out; \
	else \
		sed -i '/test/d;/commands/d;/cmd/d;/lambda/d' coverage.out; \
	fi
endef

define GENERATE_HTML
	if [ "$(HTML_FLAG)" = "html" ]; then \
		go tool cover -html=coverage.out -o coverage.html && \
		echo 'Coverage report generated: coverage.html'; \
	fi
endef

get_expo_apps:
ifeq ($(DOCKER_FLAG),docker)
	docker-compose exec ota-server go run ./internal/commands/GetExpoApplications.go
else
	go run ./commands/GetExpoApplications.go
endif

.PHONY: docker html
