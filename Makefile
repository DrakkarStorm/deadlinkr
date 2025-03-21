# Makefile pour DeadLinkr

# Variables de configuration
BINARY_NAME=deadlinkr
GO=go
GOLANGCI_LINT=golangci-lint

# Répertoires
SRC_DIR=.
BUILD_DIR=./build
COVERAGE_DIR=./coverage

# Options de compilation
LDFLAGS=-ldflags "-s -w"

# Cible par défaut
.PHONY: all
all: clean lint test build

# Nettoyer les artefacts de build
.PHONY: clean
clean:
	@echo "🧹 Nettoyage des artefacts de build..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(COVERAGE_DIR)
	@$(GO) clean -cache
	@rm -f coverage.out

# Vérifier les dépendances
.PHONY: deps
deps:
	@echo "📦 Vérification et téléchargement des dépendances..."
	@$(GO) mod tidy
	@$(GO) mod verify

# Installer les outils de développement
.PHONY: tools
tools:
	@echo "🛠️ Installation des outils de développement..."
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@$(GO) install github.com/axw/gocov/gocov@latest
	@$(GO) install github.com/matm/gocov-html/cmd/gocov-html@latest

# Linter
.PHONY: lint
lint:
	@echo "🕵️ Analyse statique du code..."
	@$(GOLANGCI_LINT) run ./...

# Tests
.PHONY: test
test:
	@echo "🚀 Déploiement de la documentation..."
	@$(GO) run ./tests/html_static.go
	@echo "⚡︎ Démarrage du serveur de documentation sur http://localhost:8085"
	@echo "🧪 Exécution des tests..."
	@mkdir -p $(COVERAGE_DIR)
	@$(GO) test -v -covermode=atomic -coverprofile=$(COVERAGE_DIR)/coverage.out ./...

# Générer le rapport de couverture
.PHONY: coverage
coverage: test
	@echo "📊 Génération du rapport de couverture..."
	@$(GO) tool cover -func=$(COVERAGE_DIR)/coverage.out
	@gocov convert $(COVERAGE_DIR)/coverage.out | gocov-html > $(COVERAGE_DIR)/coverage.html
	@echo "Rapport de couverture généré dans $(COVERAGE_DIR)/coverage.html"

# Construire le binaire
.PHONY: build
build:
	@echo "🏗️ Construction du binaire..."
	@mkdir -p $(BUILD_DIR)
	@$(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(SRC_DIR)
	@echo "Binaire construit dans $(BUILD_DIR)/$(BINARY_NAME)"

# Construire pour différentes plateformes
.PHONY: build-all
build-all:
	@echo "🌍 Construction pour toutes les plateformes..."
	@mkdir -p $(BUILD_DIR)

	# Linux
	@GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(SRC_DIR)
	@GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(SRC_DIR)

	# macOS
	@GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(SRC_DIR)
	@GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(SRC_DIR)

	# Windows
	@GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(SRC_DIR)

	@echo "Binaires construits dans $(BUILD_DIR)"

# Installation
.PHONY: install
install: build
	@echo "📥 Installation du binaire..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "DeadLinkr installé dans /usr/local/bin"

# Désinstallation
.PHONY: uninstall
uninstall:
	@echo "🗑️ Désinstallation..."
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "DeadLinkr désinstallé"

# Aide
.PHONY: help
help:
	@echo "🚀 DeadLinkr - Makefile Commandes disponibles:"
	@echo "  make            - Nettoie, lint, teste et construit le projet"
	@echo "  make deps       - Vérifie les dépendances"
	@echo "  make tools      - Installe les outils de développement"
	@echo "  make lint       - Analyse statique du code"
	@echo "  make test       - Exécute les tests"
	@echo "  make coverage   - Génère le rapport de couverture de tests"
	@echo "  make build      - Construit le binaire"
	@echo "  make build-all  - Construit pour toutes les plateformes"
	@echo "  make install    - Installe le binaire"
	@echo "  make uninstall  - Désinstalle le binaire"
	@echo "  make clean      - Nettoie les artefacts de build"
	@echo "  make help       - Affiche cette aide"

# Défaut
.DEFAULT_GOAL := help