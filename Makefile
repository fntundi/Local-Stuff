.DEFAULT_GOAL := help
.PHONY: help preflight submodules envs \
        up down destroy status \
        traefik traefik-down \
        authentik authentik-down \
        awx awx-down awx-configure \
        aistack aistack-down \
        homecam homecam-down \
        sentinel-home sentinel-home-down

SHELL   := /bin/bash
REPO    := $(shell pwd)

# AWX-as-code parameters — override on CLI: make awx-configure AWX_PASSWORD=secret
AWX_HOST     ?= http://localhost:8052
AWX_USER     ?= admin
AWX_PASSWORD ?=

# ─── Colour helpers ──────────────────────────────────────────────────────────
CYAN  := \033[36m
RESET := \033[0m
BOLD  := \033[1m
GREEN := \033[32m
YELLOW := \033[33m
RED   := \033[31m

# ─────────────────────────────────────────────────────────────────────────────

##@ Help

help: ## Show this help message
	@awk 'BEGIN {FS = ":.*##"; printf "\n$(BOLD)Home Platform — Initial Stack Setup$(RESET)\n\n$(BOLD)Usage:$(RESET)\n  make $(CYAN)<target>$(RESET)\n"} \
	  /^[a-zA-Z_-]+:.*?##/ { printf "  $(CYAN)%-24s$(RESET) %s\n", $$1, $$2 } \
	  /^##@/ { printf "\n$(BOLD)%s$(RESET)\n", substr($$0, 5) }' $(MAKEFILE_LIST)
	@echo ""

##@ Prerequisites

preflight: ## Verify all required tools are installed
	@echo "Checking prerequisites..."
	@missing=0; \
	for tool in docker git curl; do \
	  command -v $$tool >/dev/null 2>&1 \
	    && printf "  $(GREEN)✔$(RESET)  %-18s found\n" $$tool \
	    || { printf "  $(RED)✘$(RESET)  %-18s MISSING\n" $$tool; missing=1; }; \
	done; \
	docker compose version >/dev/null 2>&1 \
	  && printf "  $(GREEN)✔$(RESET)  %-18s found\n" "docker compose" \
	  || { printf "  $(RED)✘$(RESET)  %-18s MISSING\n" "docker compose"; missing=1; }; \
	for tool in k3d kubectl; do \
	  command -v $$tool >/dev/null 2>&1 \
	    && printf "  $(GREEN)✔$(RESET)  %-18s found\n" $$tool \
	    || { printf "  $(RED)✘$(RESET)  %-18s MISSING\n" $$tool; missing=1; }; \
	done; \
	command -v ansible-playbook >/dev/null 2>&1 \
	  && printf "  $(GREEN)✔$(RESET)  %-18s found\n" "ansible-playbook" \
	  || printf "  $(YELLOW)!$(RESET)  %-18s optional — needed for awx-configure\n" "ansible-playbook"; \
	[ $$missing -eq 0 ] || { echo ""; echo "Install missing tools then retry."; exit 1; }

submodules: ## Initialise and update all git submodules
	@echo "Updating submodules..."
	git submodule update --init --recursive
	@echo "  Done."

envs: ## Scaffold .env files from .env.example (skips existing files)
	@echo "Scaffolding .env files..."
	@for pair in "authentik/.env.example:authentik/.env" "awx/.env.example:awx/.env"; do \
	  src=$${pair%%:*}; dst=$${pair##*:}; \
	  if [ -f "$$dst" ]; then \
	    printf "  $(YELLOW)SKIP$(RESET)   $$dst (already exists)\n"; \
	  else \
	    cp "$$src" "$$dst"; \
	    printf "  $(GREEN)CREATE$(RESET) $$dst — fill in secrets before deploying\n"; \
	  fi; \
	done
	@if [ -f local-aistack/env/.env.prod ]; then \
	  printf "  $(YELLOW)SKIP$(RESET)   local-aistack/env/.env.prod (already exists)\n"; \
	elif [ -f local-aistack/env/.env.prod.example ]; then \
	  cp local-aistack/env/.env.prod.example local-aistack/env/.env.prod; \
	  printf "  $(GREEN)CREATE$(RESET) local-aistack/env/.env.prod — fill in secrets\n"; \
	else \
	  printf "  $(YELLOW)WARN$(RESET)   local-aistack/env/.env.prod.example not found — create manually\n"; \
	fi

##@ Full Stack

up: preflight submodules traefik authentik awx aistack homecam sentinel-home ## Deploy the full stack in dependency order
	@echo ""
	@echo "$(BOLD)$(GREEN)Stack is up.$(RESET) Run 'make status' to verify all services."
	@echo ""

down: traefik-down authentik-down awx-down aistack-down homecam-down sentinel-home-down ## Stop all services (data volumes preserved)
	@echo "All services stopped."

destroy: ## Tear down everything — removes containers, volumes, and k3d clusters
	@echo "$(BOLD)$(RED)Destroying all services and data...$(RESET)"
	-docker compose --project-name traefik-gw  -f traefik/docker-compose.yml   down -v 2>/dev/null || true
	-docker compose --project-name authentik   -f authentik/docker-compose.yml  down -v 2>/dev/null || true
	-docker compose --project-name awx         -f awx/docker-compose.yml        down -v 2>/dev/null || true
	-$(MAKE) -C local-aistack down 2>/dev/null || true
	-k3d cluster delete sentinel-noc  2>/dev/null || true
	-k3d cluster delete sentinel-home 2>/dev/null || true
	@echo "Destroy complete."

##@ Services — Docker Compose (home-server)

traefik: ## Deploy Traefik reverse proxy (creates home-net 172.20.0.0/16, TLS termination)
	@echo "$(CYAN)▶  Traefik$(RESET)"
	docker compose --project-name traefik-gw -f traefik/docker-compose.yml up -d
	@echo -n "  Waiting for Traefik to accept connections..."
	@for i in $$(seq 1 30); do \
	  curl -s --connect-timeout 2 http://localhost:80 -o /dev/null 2>/dev/null && { echo " ready"; exit 0; }; \
	  [ $$i -eq 30 ] && { echo ""; echo "  $(RED)Timeout.$(RESET) Check: docker compose -p traefik-gw logs"; exit 1; }; \
	  sleep 2; \
	done

traefik-down: ## Stop Traefik
	-docker compose --project-name traefik-gw -f traefik/docker-compose.yml down 2>/dev/null || true

authentik: ## Deploy Authentik SSO (requires home-net from Traefik)
	@echo "$(CYAN)▶  Authentik$(RESET)"
	@[ -f authentik/.env ] || { \
	  echo "  $(RED)[ERROR]$(RESET) authentik/.env not found — run 'make envs' first"; exit 1; }
	docker compose --project-name authentik -f authentik/docker-compose.yml up -d
	@echo -n "  Waiting for Authentik health endpoint (~60 s)..."
	@for i in $$(seq 1 60); do \
	  curl -sf --connect-timeout 2 http://localhost:9000/-/health/live/ -o /dev/null 2>/dev/null \
	    && { echo " ready"; exit 0; }; \
	  [ $$i -eq 60 ] && { echo ""; \
	    echo "  $(RED)Timeout.$(RESET) Check: docker compose -p authentik logs"; exit 1; }; \
	  sleep 3; \
	done

authentik-down: ## Stop Authentik
	-docker compose --project-name authentik -f authentik/docker-compose.yml down 2>/dev/null || true

awx: ## Deploy AWX automation controller (~2 min startup)
	@echo "$(CYAN)▶  AWX$(RESET)"
	@[ -f awx/.env ] || { \
	  echo "  $(RED)[ERROR]$(RESET) awx/.env not found — run 'make envs' first"; exit 1; }
	docker compose --project-name awx -f awx/docker-compose.yml up -d
	@echo -n "  Waiting for AWX API (may take ~2 min)..."
	@for i in $$(seq 1 60); do \
	  curl -sf --connect-timeout 3 http://localhost:8052/api/v2/ping/ -o /dev/null 2>/dev/null \
	    && { echo " ready"; exit 0; }; \
	  [ $$i -eq 60 ] && { echo ""; \
	    echo "  $(RED)Timeout.$(RESET) Check: docker compose -p awx logs"; exit 1; }; \
	  sleep 5; \
	done

awx-down: ## Stop AWX
	-docker compose --project-name awx -f awx/docker-compose.yml down 2>/dev/null || true

awx-configure: ## Apply AWX-as-code configuration (requires AWX_PASSWORD=<password>)
	@[ -n "$(AWX_PASSWORD)" ] || { \
	  echo "  $(RED)[ERROR]$(RESET) AWX_PASSWORD is required:"; \
	  echo "          make awx-configure AWX_PASSWORD=<your-admin-password>"; \
	  exit 1; }
	@command -v ansible-playbook >/dev/null 2>&1 || { \
	  echo "  $(RED)[ERROR]$(RESET) ansible-playbook not found — install ansible first"; exit 1; }
	ansible-playbook awx/awx_config/configure_awx.yml \
	  -e "awx_host=$(AWX_HOST)" \
	  -e "awx_username=$(AWX_USER)" \
	  -e "awx_password=$(AWX_PASSWORD)"

##@ Services — AI Platform (local-aistack submodule)

# Must delegate to submodule Makefile — compose bind-mounts use ${PWD} which
# must resolve to local-aistack/, not the repo root.
aistack: ## Deploy local-aistack AI platform (Ollama, OpenWebUI, LiteLLM, MinIO, ...)
	@echo "$(CYAN)▶  AI Platform (local-aistack)$(RESET)"
	@[ -d local-aistack ] || { \
	  echo "  $(RED)[ERROR]$(RESET) local-aistack/ not found — run 'make submodules'"; exit 1; }
	@[ -f local-aistack/env/.env.prod ] || { \
	  echo "  $(RED)[ERROR]$(RESET) local-aistack/env/.env.prod not found — run 'make envs' first"; exit 1; }
	$(MAKE) -C local-aistack init
	$(MAKE) -C local-aistack up

aistack-down: ## Stop local-aistack AI platform
	-$(MAKE) -C local-aistack down 2>/dev/null || true

##@ Services — HomeCam (k3d cluster: sentinel-noc)

homecam: ## Deploy HomeCam NOC to k3d cluster sentinel-noc (creates ai-home-shared 172.30.0.0/24)
	@echo "$(CYAN)▶  HomeCam (k3d: sentinel-noc)$(RESET)"
	@[ -d HomeCam ] || { \
	  echo "  $(RED)[ERROR]$(RESET) HomeCam/ not found — run 'make submodules'"; exit 1; }
	$(MAKE) -C HomeCam k8s-up
	@echo "  HomeCam deployed."
	@echo "  Frontend NodePort : http://<host>:30080"
	@echo "  API NodePort      : http://<host>:30081"

homecam-down: ## Delete HomeCam k3d cluster sentinel-noc
	-k3d cluster delete sentinel-noc 2>/dev/null || true

##@ Services — Sentinel-Home (k3d cluster: sentinel-home)

sentinel-home: ## Deploy Sentinel-Home to k3d cluster sentinel-home (network: home-net)
	@echo "$(CYAN)▶  Sentinel-Home (k3d: sentinel-home)$(RESET)"
	@[ -d sentinel-home ] || { \
	  echo "  $(RED)[ERROR]$(RESET) sentinel-home/ not found — run 'make submodules'"; exit 1; }
	@[ -f infra/sentinel-home/k3d-config.yaml ] || { \
	  echo "  $(RED)[ERROR]$(RESET) infra/sentinel-home/k3d-config.yaml not found"; exit 1; }
	@if k3d cluster list 2>/dev/null | grep -q 'sentinel-home'; then \
	  echo "  Cluster sentinel-home already exists — skipping create."; \
	else \
	  k3d cluster create --config infra/sentinel-home/k3d-config.yaml; \
	fi
	@manifest_dir=sentinel-home/deploy/kubernetes; \
	if [ -d "$$manifest_dir" ] && ls "$$manifest_dir"/*.yaml >/dev/null 2>&1; then \
	  kubectl --context k3d-sentinel-home apply -f "$$manifest_dir/"; \
	  echo "  Sentinel-Home deployed."; \
	  echo "  Frontend NodePort : http://<host>:31000"; \
	  echo "  API NodePort      : http://<host>:31001"; \
	else \
	  echo "  $(YELLOW)[WARN]$(RESET)  No manifests found in $$manifest_dir — cluster created, skipping apply."; \
	  echo "         Sentinel-Home is still in development; deploy manifests manually when ready."; \
	fi

sentinel-home-down: ## Delete Sentinel-Home k3d cluster
	-k3d cluster delete sentinel-home 2>/dev/null || true

##@ Observability

status: ## Show current health status for all home-server services and k3d clusters
	@echo ""
	@echo "$(BOLD)Service Status$(RESET)"
	@echo "──────────────────────────────────────────────────"
	@_status() { \
	  name=$$1; url=$$2; \
	  if curl -s --connect-timeout 3 "$$url" -o /dev/null 2>/dev/null; then \
	    printf "  $(GREEN)UP$(RESET)      %-20s  %s\n" "$$name" "$$url"; \
	  else \
	    printf "  $(RED)DOWN$(RESET)    %-20s  %s\n" "$$name" "$$url"; \
	  fi; \
	}; \
	_status "Traefik"   "http://localhost:80"; \
	_status "Authentik" "http://localhost:9000/-/health/live/"; \
	_status "AWX"       "http://localhost:8052/api/v2/ping/"
	@echo ""
	@echo "$(BOLD)Docker Compose stacks$(RESET)"
	@echo "──────────────────────────────────────────────────"
	@for proj in traefik-gw authentik awx ai-platform; do \
	  running=$$(docker compose -p $$proj ps --services --filter status=running 2>/dev/null | wc -l | tr -d ' '); \
	  total=$$(docker compose -p $$proj ps --services 2>/dev/null | wc -l | tr -d ' '); \
	  if [ "$$total" -gt 0 ] 2>/dev/null; then \
	    printf "  %-16s  %s/%s containers running\n" "$$proj" "$$running" "$$total"; \
	  else \
	    printf "  %-16s  $(YELLOW)not deployed$(RESET)\n" "$$proj"; \
	  fi; \
	done
	@echo ""
	@echo "$(BOLD)k3d clusters$(RESET)"
	@echo "──────────────────────────────────────────────────"
	@k3d cluster list 2>/dev/null || echo "  k3d not available"
	@echo ""
