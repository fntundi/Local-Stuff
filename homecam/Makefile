# Sentinel NOC - Makefile
# =====================
# Manage Docker Compose (dev) and Kubernetes (k3d) deployments.
# Run `make help` to see all available targets.

.PHONY: help build up down restart status logs logs-backend logs-frontend logs-db logs-mediamtx \
        shell-backend shell-frontend shell-db test test-backend test-frontend \
        lint lint-backend lint-frontend clean clean-images prune reset-db backup-db restore-db \
        dev install backend-tidy frontend-lock version health \
        k3d-create k3d-delete k3d-list k3d-start k3d-stop \
        k8s-build k8s-push k8s-secrets k8s-deploy k8s-undeploy k8s-redeploy \
        k8s-status k8s-health k8s-rollout k8s-logs k8s-logs-backend k8s-logs-frontend \
        k8s-logs-db k8s-logs-mediamtx k8s-port-forward \
        k8s-shell-backend k8s-shell-frontend k8s-shell-db \
        k8s-scale-backend k8s-scale-frontend k8s-up k8s-down

.DEFAULT_GOAL := help

# Colors
COLOR_RESET   = \033[0m
COLOR_INFO    = \033[36m
COLOR_SUCCESS = \033[32m
COLOR_WARNING = \033[33m
COLOR_ERROR   = \033[31m

DOCKER_COMPOSE = docker-compose
K3D_CLUSTER    = sentinel-noc
K8S_NAMESPACE  = sentinel-noc
K8S_DIR        = ./k8s

# =============================================================================
# HELP
# =============================================================================

help: ## Show this help message
	@echo ""
	@echo "$(COLOR_INFO)Sentinel NOC - Security Camera Network Operations Center$(COLOR_RESET)"
	@echo "$(COLOR_INFO)========================================================$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_WARNING)Usage:$(COLOR_RESET) make [target]"
	@echo ""
	@echo "$(COLOR_SUCCESS)Docker Compose (local dev):$(COLOR_RESET)"
	@grep -E '^(build|up|down|restart|status|logs|shell|test|lint|clean|dev|install|version|health|reset-db|backup-db|restore-db|prune)[^:]*:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_INFO)%-28s$(COLOR_RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(COLOR_SUCCESS)Kubernetes - Cluster (k3d):$(COLOR_RESET)"
	@grep -E '^k3d[^:]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_INFO)%-28s$(COLOR_RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(COLOR_SUCCESS)Kubernetes - Images:$(COLOR_RESET)"
	@grep -E '^k8s-(build|push|secrets)[^:]*:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_INFO)%-28s$(COLOR_RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(COLOR_SUCCESS)Kubernetes - Deployment:$(COLOR_RESET)"
	@grep -E '^k8s-(deploy|undeploy|redeploy|status|health|rollout|up|down)[^:]*:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_INFO)%-28s$(COLOR_RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(COLOR_SUCCESS)Kubernetes - Operations:$(COLOR_RESET)"
	@grep -E '^k8s-(logs|port-forward|shell|scale)[^:]*:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_INFO)%-28s$(COLOR_RESET) %s\n", $$1, $$2}'
	@echo ""

# =============================================================================
# DOCKER COMPOSE - BUILD & RUN
# =============================================================================

build: ## Build all Docker images
	@echo "$(COLOR_INFO)Building Docker images...$(COLOR_RESET)"
	$(DOCKER_COMPOSE) build
	@echo "$(COLOR_SUCCESS)Build complete!$(COLOR_RESET)"

build-no-cache: ## Build all Docker images without cache
	@echo "$(COLOR_INFO)Building Docker images (no cache)...$(COLOR_RESET)"
	$(DOCKER_COMPOSE) build --no-cache
	@echo "$(COLOR_SUCCESS)Build complete!$(COLOR_RESET)"

up: ## Start all services in detached mode
	@echo "$(COLOR_INFO)Starting services...$(COLOR_RESET)"
	$(DOCKER_COMPOSE) up -d
	@echo "$(COLOR_SUCCESS)Services started!$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_INFO)Access the application:$(COLOR_RESET)"
	@echo "  Frontend:  http://localhost:3000"
	@echo "  Backend:   http://localhost:8001/api"
	@echo "  HLS:       http://localhost:8888"
	@echo "  RTSP:      rtsp://localhost:8554"
	@echo ""
	@echo "$(COLOR_WARNING)Default login:$(COLOR_RESET) admin / P@ssw0rd!"

up-attached: ## Start all services in attached mode (see logs)
	@echo "$(COLOR_INFO)Starting services (attached)...$(COLOR_RESET)"
	$(DOCKER_COMPOSE) up

down: ## Stop all services
	@echo "$(COLOR_INFO)Stopping services...$(COLOR_RESET)"
	$(DOCKER_COMPOSE) down
	@echo "$(COLOR_SUCCESS)Services stopped!$(COLOR_RESET)"

restart: ## Restart all services
	@echo "$(COLOR_INFO)Restarting services...$(COLOR_RESET)"
	$(DOCKER_COMPOSE) restart
	@echo "$(COLOR_SUCCESS)Services restarted!$(COLOR_RESET)"

restart-backend: ## Restart backend service only
	@echo "$(COLOR_INFO)Restarting backend...$(COLOR_RESET)"
	$(DOCKER_COMPOSE) restart backend
	@echo "$(COLOR_SUCCESS)Backend restarted!$(COLOR_RESET)"

restart-frontend: ## Restart frontend service only
	@echo "$(COLOR_INFO)Restarting frontend...$(COLOR_RESET)"
	$(DOCKER_COMPOSE) restart frontend
	@echo "$(COLOR_SUCCESS)Frontend restarted!$(COLOR_RESET)"

status: ## Show status of all services
	@echo "$(COLOR_INFO)Service Status:$(COLOR_RESET)"
	@echo ""
	$(DOCKER_COMPOSE) ps

# =============================================================================
# LOGS
# =============================================================================

logs: ## View logs from all services (follow mode)
	$(DOCKER_COMPOSE) logs -f

logs-backend: ## View backend logs (follow mode)
	$(DOCKER_COMPOSE) logs -f backend

logs-frontend: ## View frontend logs (follow mode)
	$(DOCKER_COMPOSE) logs -f frontend

logs-db: ## View MongoDB logs (follow mode)
	$(DOCKER_COMPOSE) logs -f mongodb

logs-mediamtx: ## View MediaMTX streaming server logs (follow mode)
	$(DOCKER_COMPOSE) logs -f mediamtx

logs-tail: ## View last 100 lines of all logs
	$(DOCKER_COMPOSE) logs --tail=100

# =============================================================================
# SHELL ACCESS
# =============================================================================

shell-backend: ## Open a shell in the backend container
	@echo "$(COLOR_INFO)Opening backend shell...$(COLOR_RESET)"
	$(DOCKER_COMPOSE) exec backend /bin/sh

shell-frontend: ## Open a shell in the frontend container
	@echo "$(COLOR_INFO)Opening frontend shell...$(COLOR_RESET)"
	$(DOCKER_COMPOSE) exec frontend /bin/sh

shell-db: ## Open MongoDB shell
	@echo "$(COLOR_INFO)Opening MongoDB shell...$(COLOR_RESET)"
	$(DOCKER_COMPOSE) exec mongodb mongosh sentinel_noc

# =============================================================================
# TESTING
# =============================================================================

test: test-backend test-frontend ## Run all tests
	@echo "$(COLOR_SUCCESS)All tests complete!$(COLOR_RESET)"

test-backend: ## Run Go backend tests
	@echo "$(COLOR_INFO)Running backend tests...$(COLOR_RESET)"
	$(DOCKER_COMPOSE) exec backend go test ./... -v

test-frontend: ## Run frontend tests
	@echo "$(COLOR_INFO)Running frontend tests...$(COLOR_RESET)"
	$(DOCKER_COMPOSE) exec frontend yarn test --watchAll=false

# =============================================================================
# LINTING
# =============================================================================

lint: lint-backend lint-frontend ## Run all linters
	@echo "$(COLOR_SUCCESS)Linting complete!$(COLOR_RESET)"

lint-backend: ## Lint Go backend code
	@echo "$(COLOR_INFO)Linting backend...$(COLOR_RESET)"
	$(DOCKER_COMPOSE) exec backend go vet ./...

lint-frontend: ## Lint frontend code
	@echo "$(COLOR_INFO)Linting frontend...$(COLOR_RESET)"
	$(DOCKER_COMPOSE) exec frontend yarn lint

# =============================================================================
# DATABASE
# =============================================================================

reset-db: ## Reset database (WARNING: deletes all data)
	@echo "$(COLOR_WARNING)WARNING: This will delete all data!$(COLOR_RESET)"
	@read -p "Are you sure? [y/N] " confirm && [ "$$confirm" = "y" ] || exit 1
	@echo "$(COLOR_INFO)Resetting database...$(COLOR_RESET)"
	$(DOCKER_COMPOSE) exec mongodb mongosh sentinel_noc --eval "db.dropDatabase()"
	$(DOCKER_COMPOSE) restart backend
	@echo "$(COLOR_SUCCESS)Database reset complete!$(COLOR_RESET)"

backup-db: ## Backup database to ./backups directory
	@echo "$(COLOR_INFO)Backing up database...$(COLOR_RESET)"
	@mkdir -p ./backups
	$(DOCKER_COMPOSE) exec mongodb mongodump --db sentinel_noc --out /tmp/backup
	$(DOCKER_COMPOSE) cp mongodb:/tmp/backup ./backups/backup-$$(date +%Y%m%d-%H%M%S)
	@echo "$(COLOR_SUCCESS)Backup complete!$(COLOR_RESET)"

restore-db: ## Restore database from backup (usage: make restore-db BACKUP=./backups/backup-xxx)
	@echo "$(COLOR_INFO)Restoring database from $(BACKUP)...$(COLOR_RESET)"
	$(DOCKER_COMPOSE) cp $(BACKUP) mongodb:/tmp/restore
	$(DOCKER_COMPOSE) exec mongodb mongorestore --db sentinel_noc /tmp/restore/sentinel_noc
	@echo "$(COLOR_SUCCESS)Restore complete!$(COLOR_RESET)"

# =============================================================================
# DEVELOPMENT
# =============================================================================

dev: ## Start in development mode (hot reload)
	@echo "$(COLOR_INFO)Starting in development mode...$(COLOR_RESET)"
	$(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.dev.yml up

install: ## Install dependencies locally (for development without Docker)
	@echo "$(COLOR_INFO)Installing backend dependencies...$(COLOR_RESET)"
	cd backend && go mod download
	@echo "$(COLOR_INFO)Installing frontend dependencies...$(COLOR_RESET)"
	cd frontend && yarn install
	@echo "$(COLOR_SUCCESS)Dependencies installed!$(COLOR_RESET)"

backend-tidy: ## Generate backend/go.sum via Docker if it does not exist (commit the result)
	@if [ -f ./backend/go.sum ]; then \
		echo "$(COLOR_INFO)backend/go.sum already exists — skipping generation.$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_INFO)Generating backend/go.sum via Docker (golang:1.21-alpine)...$(COLOR_RESET)"; \
		docker run --rm \
			-v "$$(pwd)/backend":/app \
			-w /app \
			golang:1.21-alpine \
			sh -c "go mod download"; \
		echo "$(COLOR_SUCCESS)go.sum generated. Commit it: git add backend/go.sum$(COLOR_RESET)"; \
	fi

frontend-lock: ## Generate frontend/yarn.lock via Docker if it does not exist (commit the result)
	@if [ -f ./frontend/yarn.lock ]; then \
		echo "$(COLOR_INFO)frontend/yarn.lock already exists — skipping generation.$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_INFO)Generating frontend/yarn.lock via Docker (node:20-alpine)...$(COLOR_RESET)"; \
		docker run --rm \
			-v "$$(pwd)/frontend":/app \
			-w /app \
			node:20-alpine \
			sh -c "yarn install --no-progress 2>&1"; \
		echo "$(COLOR_SUCCESS)yarn.lock generated. Commit it: git add frontend/yarn.lock$(COLOR_RESET)"; \
	fi

# =============================================================================
# CLEANUP
# =============================================================================

clean: ## Remove containers, networks, and volumes
	@echo "$(COLOR_WARNING)WARNING: This will remove all containers and volumes!$(COLOR_RESET)"
	@read -p "Are you sure? [y/N] " confirm && [ "$$confirm" = "y" ] || exit 1
	@echo "$(COLOR_INFO)Cleaning up...$(COLOR_RESET)"
	$(DOCKER_COMPOSE) down -v --remove-orphans
	@echo "$(COLOR_SUCCESS)Cleanup complete!$(COLOR_RESET)"

clean-images: ## Remove all project Docker images
	@echo "$(COLOR_INFO)Removing Docker images...$(COLOR_RESET)"
	$(DOCKER_COMPOSE) down --rmi all
	@echo "$(COLOR_SUCCESS)Images removed!$(COLOR_RESET)"

prune: ## Remove all unused Docker resources
	@echo "$(COLOR_WARNING)WARNING: This will remove all unused Docker resources!$(COLOR_RESET)"
	@read -p "Are you sure? [y/N] " confirm && [ "$$confirm" = "y" ] || exit 1
	docker system prune -af
	@echo "$(COLOR_SUCCESS)Prune complete!$(COLOR_RESET)"

# =============================================================================
# UTILITY
# =============================================================================

ps: status ## Alias for status

version: ## Show version information
	@echo "$(COLOR_INFO)Sentinel NOC$(COLOR_RESET)"
	@echo "Version: 1.0.0"
	@echo ""
	@echo "Docker Compose version:"
	@$(DOCKER_COMPOSE) version

health: ## Check health of all services (Docker Compose)
	@echo "$(COLOR_INFO)Checking service health...$(COLOR_RESET)"
	@echo ""
	@echo "Backend API:"
	@curl -s http://localhost:8001/api/health | python3 -m json.tool 2>/dev/null || echo "$(COLOR_ERROR)Backend not responding$(COLOR_RESET)"
	@echo ""
	@echo "Frontend:"
	@curl -s -o /dev/null -w "Status: %{http_code}" http://localhost:3000 && echo " $(COLOR_SUCCESS)OK$(COLOR_RESET)" || echo "$(COLOR_ERROR)Frontend not responding$(COLOR_RESET)"
	@echo ""
	@echo "MediaMTX:"
	@curl -s http://localhost:9997/v3/paths/list > /dev/null 2>&1 && echo "Status: $(COLOR_SUCCESS)OK$(COLOR_RESET)" || echo "$(COLOR_ERROR)MediaMTX not responding$(COLOR_RESET)"
	@echo ""
	@echo "MongoDB:"
	@$(DOCKER_COMPOSE) exec -T mongodb mongosh --eval "db.adminCommand('ping')" > /dev/null 2>&1 && echo "Status: $(COLOR_SUCCESS)OK$(COLOR_RESET)" || echo "$(COLOR_ERROR)MongoDB not responding$(COLOR_RESET)"

# =============================================================================
# KUBERNETES - CLUSTER MANAGEMENT (k3d)
# =============================================================================

k3d-create: ## Create local k3d Kubernetes cluster (sentinel-noc); no-op if already exists
	@echo "$(COLOR_INFO)Ensuring shared Docker network ai-home-shared exists...$(COLOR_RESET)"
	@docker network create ai-home-shared \
		--driver bridge --subnet 172.30.0.0/24 \
		2>/dev/null || true
	@if k3d cluster list -o json 2>/dev/null | grep -q '"name":"$(K3D_CLUSTER)"'; then \
		echo "$(COLOR_WARNING)Cluster '$(K3D_CLUSTER)' already exists — skipping creation.$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_INFO)Creating k3d cluster '$(K3D_CLUSTER)'...$(COLOR_RESET)"; \
		k3d cluster create --config $(K8S_DIR)/k3d-config.yaml; \
		echo "$(COLOR_SUCCESS)Cluster created!$(COLOR_RESET)"; \
	fi
	@echo "$(COLOR_SUCCESS)Cluster ready. HTTP routing is handled by Traefik in the ai-platform cluster.$(COLOR_RESET)"
	@echo "$(COLOR_INFO)Run: cd ../local-aistack/hybrid-ai-platform && make homecam-connect$(COLOR_RESET)"

k3d-delete: ## Delete the k3d cluster; no-op if it does not exist
	@if k3d cluster list -o json 2>/dev/null | grep -q '"name":"$(K3D_CLUSTER)"'; then \
		echo "$(COLOR_WARNING)Deleting k3d cluster '$(K3D_CLUSTER)'...$(COLOR_RESET)"; \
		k3d cluster delete $(K3D_CLUSTER); \
		echo "$(COLOR_SUCCESS)Cluster deleted.$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_WARNING)Cluster '$(K3D_CLUSTER)' does not exist — nothing to delete.$(COLOR_RESET)"; \
	fi

k3d-list: ## List all k3d clusters
	k3d cluster list

k3d-start: ## Start a stopped k3d cluster
	@echo "$(COLOR_INFO)Starting k3d cluster '$(K3D_CLUSTER)'...$(COLOR_RESET)"
	k3d cluster start $(K3D_CLUSTER)
	@echo "$(COLOR_SUCCESS)Cluster started.$(COLOR_RESET)"

k3d-stop: ## Stop the k3d cluster (without deleting it)
	@echo "$(COLOR_INFO)Stopping k3d cluster '$(K3D_CLUSTER)'...$(COLOR_RESET)"
	k3d cluster stop $(K3D_CLUSTER)
	@echo "$(COLOR_SUCCESS)Cluster stopped.$(COLOR_RESET)"

# =============================================================================
# KUBERNETES - IMAGE MANAGEMENT
# =============================================================================

k8s-build: backend-tidy frontend-lock ## Build Docker images tagged for Kubernetes deployment
	@echo "$(COLOR_INFO)Building Docker images for Kubernetes...$(COLOR_RESET)"
	docker build -t sentinel-backend:latest ./backend
	docker build -t sentinel-frontend:latest -f ./frontend/Dockerfile.prod ./frontend
	@echo "$(COLOR_SUCCESS)Images built!$(COLOR_RESET)"

k8s-push: ## Import locally built images into the k3d cluster
	@echo "$(COLOR_INFO)Importing images into k3d cluster '$(K3D_CLUSTER)'...$(COLOR_RESET)"
	k3d image import sentinel-backend:latest sentinel-frontend:latest -c $(K3D_CLUSTER)
	@echo "$(COLOR_SUCCESS)Images imported!$(COLOR_RESET)"

k8s-secrets: ## Generate and apply cryptographically secure secrets
	@echo "$(COLOR_INFO)Generating Kubernetes secrets...$(COLOR_RESET)"
	@JWT_SECRET=$$(openssl rand -base64 48) && \
	ENC_KEY=$$(openssl rand -base64 32 | head -c 44)= && \
	MONGO_PASS=$$(openssl rand -base64 24) && \
	kubectl create secret generic sentinel-secrets \
		--namespace $(K8S_NAMESPACE) \
		--from-literal=jwt-secret="$$JWT_SECRET" \
		--from-literal=encryption-key="$$ENC_KEY" \
		--from-literal=mongo-username="admin" \
		--from-literal=mongo-password="$$MONGO_PASS" \
		--dry-run=client -o yaml | kubectl apply -f -
	@echo "$(COLOR_SUCCESS)Secrets applied!$(COLOR_RESET)"

# =============================================================================
# KUBERNETES - DEPLOYMENT
# =============================================================================

k8s-deploy: ## Deploy all application resources to Kubernetes
	@echo "$(COLOR_INFO)Deploying Sentinel NOC to Kubernetes...$(COLOR_RESET)"
	kubectl apply -f $(K8S_DIR)/namespace.yaml
	kubectl apply -f $(K8S_DIR)/configmap.yaml
	kubectl apply -f $(K8S_DIR)/secrets.yaml
	kubectl apply -f $(K8S_DIR)/mongodb/
	kubectl apply -f $(K8S_DIR)/mediamtx/
	kubectl apply -f $(K8S_DIR)/backend/
	kubectl apply -f $(K8S_DIR)/frontend/
	kubectl apply -f $(K8S_DIR)/network-policy.yaml
	@echo "$(COLOR_SUCCESS)Deployment applied!$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_INFO)Waiting for pods to be ready...$(COLOR_RESET)"
	kubectl wait --namespace $(K8S_NAMESPACE) \
		--for=condition=ready pod \
		--selector=app=backend \
		--timeout=120s 2>/dev/null || true
	@echo ""
	@$(MAKE) k8s-status

k8s-undeploy: ## Remove all application resources from Kubernetes (keeps namespace)
	@echo "$(COLOR_WARNING)Removing Sentinel NOC from Kubernetes...$(COLOR_RESET)"
	kubectl delete -f $(K8S_DIR)/network-policy.yaml --ignore-not-found
	kubectl delete -f $(K8S_DIR)/frontend/ --ignore-not-found
	kubectl delete -f $(K8S_DIR)/backend/ --ignore-not-found
	kubectl delete -f $(K8S_DIR)/mediamtx/ --ignore-not-found
	kubectl delete -f $(K8S_DIR)/mongodb/ --ignore-not-found
	@echo "$(COLOR_SUCCESS)Resources removed.$(COLOR_RESET)"

k8s-redeploy: k8s-build k8s-push k8s-rollout ## Rebuild images and restart all deployments

k8s-rollout: ## Restart all Kubernetes deployments (rolling update)
	@echo "$(COLOR_INFO)Rolling restart of all deployments...$(COLOR_RESET)"
	kubectl rollout restart deployment/backend -n $(K8S_NAMESPACE)
	kubectl rollout restart deployment/frontend -n $(K8S_NAMESPACE)
	kubectl rollout restart deployment/mediamtx -n $(K8S_NAMESPACE)
	@echo "$(COLOR_SUCCESS)Rollout triggered.$(COLOR_RESET)"

k8s-status: ## Show Kubernetes pod and service status
	@echo "$(COLOR_INFO)Pods in namespace '$(K8S_NAMESPACE)':$(COLOR_RESET)"
	kubectl get pods -n $(K8S_NAMESPACE) -o wide
	@echo ""
	@echo "$(COLOR_INFO)Services:$(COLOR_RESET)"
	kubectl get services -n $(K8S_NAMESPACE)
	@echo ""
	@echo "$(COLOR_INFO)Ingress:$(COLOR_RESET)"
	kubectl get ingress -n $(K8S_NAMESPACE)

k8s-health: ## Check Kubernetes service health via port-forward
	@echo "$(COLOR_INFO)Checking Kubernetes service health...$(COLOR_RESET)"
	@echo ""
	@echo "Backend API (via service):"
	@kubectl run --rm -i --restart=Never health-check \
		--image=curlimages/curl:latest -n $(K8S_NAMESPACE) \
		-- curl -sf http://backend-service:8001/api/health 2>/dev/null \
		|| echo "$(COLOR_ERROR)Backend health check failed$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_INFO)For external access, run: make k8s-port-forward$(COLOR_RESET)"

# =============================================================================
# KUBERNETES - OPERATIONS
# =============================================================================

k8s-logs: ## Stream logs from all pods in sentinel-noc namespace
	kubectl logs -n $(K8S_NAMESPACE) -l app.kubernetes.io/part-of=sentinel-noc -f --max-log-requests=10

k8s-logs-backend: ## Stream backend pod logs
	kubectl logs -n $(K8S_NAMESPACE) -l app=backend -f

k8s-logs-frontend: ## Stream frontend pod logs
	kubectl logs -n $(K8S_NAMESPACE) -l app=frontend -f

k8s-logs-db: ## Stream MongoDB pod logs
	kubectl logs -n $(K8S_NAMESPACE) -l app=mongodb -f

k8s-logs-mediamtx: ## Stream MediaMTX pod logs
	kubectl logs -n $(K8S_NAMESPACE) -l app=mediamtx -f

k8s-port-forward: ## Port-forward services to localhost for local access
	@echo "$(COLOR_INFO)Port-forwarding Sentinel NOC services...$(COLOR_RESET)"
	@echo "  Frontend → http://localhost:3000"
	@echo "  Backend  → http://localhost:8001"
	@echo "  MediaMTX HLS → http://localhost:8888"
	@echo "  Press Ctrl+C to stop"
	@kubectl port-forward -n $(K8S_NAMESPACE) svc/frontend-service 3000:80 &
	@kubectl port-forward -n $(K8S_NAMESPACE) svc/backend-service 8001:8001 &
	@kubectl port-forward -n $(K8S_NAMESPACE) svc/mediamtx-service 8888:8888 &
	@wait

k8s-shell-backend: ## Open a shell in the backend pod
	@echo "$(COLOR_INFO)Opening backend pod shell...$(COLOR_RESET)"
	@POD=$$(kubectl get pod -n $(K8S_NAMESPACE) -l app=backend -o jsonpath='{.items[0].metadata.name}') && \
	kubectl exec -it -n $(K8S_NAMESPACE) $$POD -- /bin/sh

k8s-shell-frontend: ## Open a shell in the frontend pod
	@echo "$(COLOR_INFO)Opening frontend pod shell...$(COLOR_RESET)"
	@POD=$$(kubectl get pod -n $(K8S_NAMESPACE) -l app=frontend -o jsonpath='{.items[0].metadata.name}') && \
	kubectl exec -it -n $(K8S_NAMESPACE) $$POD -- /bin/sh

k8s-shell-db: ## Open MongoDB shell in the pod
	@echo "$(COLOR_INFO)Opening MongoDB shell...$(COLOR_RESET)"
	@POD=$$(kubectl get pod -n $(K8S_NAMESPACE) -l app=mongodb -o jsonpath='{.items[0].metadata.name}') && \
	kubectl exec -it -n $(K8S_NAMESPACE) $$POD -- mongosh sentinel_noc \
		-u admin -p "$$MONGO_PASSWORD" --authenticationDatabase admin

# =============================================================================
# KUBERNETES - SCALING
# =============================================================================

k8s-scale-backend: ## Scale backend replicas (usage: make k8s-scale-backend REPLICAS=3)
	@echo "$(COLOR_INFO)Scaling backend to $(REPLICAS) replicas...$(COLOR_RESET)"
	kubectl scale deployment/backend -n $(K8S_NAMESPACE) --replicas=$(REPLICAS)
	@echo "$(COLOR_SUCCESS)Scaled!$(COLOR_RESET)"

k8s-scale-frontend: ## Scale frontend replicas (usage: make k8s-scale-frontend REPLICAS=3)
	@echo "$(COLOR_INFO)Scaling frontend to $(REPLICAS) replicas...$(COLOR_RESET)"
	kubectl scale deployment/frontend -n $(K8S_NAMESPACE) --replicas=$(REPLICAS)
	@echo "$(COLOR_SUCCESS)Scaled!$(COLOR_RESET)"

# =============================================================================
# KUBERNETES - ALL-IN-ONE COMMANDS
# =============================================================================

k8s-up: ## Full setup: create cluster, install ingress, build images, deploy app
	@echo "$(COLOR_INFO)Full Kubernetes setup for Sentinel NOC...$(COLOR_RESET)"
	@echo ""
	@echo "Step 1/4: Creating k3d cluster..."
	@$(MAKE) k3d-create
	@echo ""
	@echo "Step 2/4: Building Docker images..."
	@$(MAKE) k8s-build
	@echo ""
	@echo "Step 3/4: Importing images into cluster..."
	@$(MAKE) k8s-push
	@echo ""
	@echo "Step 4/4: Deploying application..."
	@$(MAKE) k8s-deploy
	@echo ""
	@echo "$(COLOR_SUCCESS)================================================================$(COLOR_RESET)"
	@echo "$(COLOR_SUCCESS) Sentinel NOC is running on Kubernetes!$(COLOR_RESET)"
	@echo "$(COLOR_SUCCESS)================================================================$(COLOR_RESET)"
	@echo ""
	@echo "  Frontend:  http://localhost"
	@echo "  Backend:   http://localhost/api"
	@echo "  HLS:       http://localhost:8888"
	@echo "  RTSP:      rtsp://localhost:8554"
	@echo ""
	@echo "$(COLOR_WARNING)Default login:$(COLOR_RESET) admin / P@ssw0rd!"
	@echo ""
	@echo "$(COLOR_WARNING)For production, run: make k8s-secrets$(COLOR_RESET)"

k8s-down: ## Tear down application and delete the k3d cluster entirely
	@echo "$(COLOR_WARNING)Tearing down Sentinel NOC Kubernetes cluster...$(COLOR_RESET)"
	@read -p "This will delete the cluster and ALL data. Are you sure? [y/N] " confirm && [ "$$confirm" = "y" ] || exit 1
	@$(MAKE) k3d-delete
	@echo "$(COLOR_SUCCESS)Cluster destroyed.$(COLOR_RESET)"
