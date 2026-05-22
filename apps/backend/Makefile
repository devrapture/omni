SHELL := /bin/sh

.PHONY: migrate-diff migrate-apply migrate-status migrate-hash

ATLAS ?= atlas
ENV ?= local

define load_env
set -a; [ ! -f .env ] || . ./.env; set +a;
endef

migrate-diff:
ifndef NAME
	$(error NAME is required. Usage: make migrate-diff NAME=add_pgvector_extension)
endif
	@$(load_env) $(ATLAS) migrate diff --env $(ENV) "$(NAME)"

migrate-apply:
	@$(load_env) $(ATLAS) migrate apply --env $(ENV)

migrate-status:
	@$(load_env) $(ATLAS) migrate status --env $(ENV)

migrate-hash:
	@$(ATLAS) migrate hash --dir file://internal/migrations
