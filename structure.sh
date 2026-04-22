#!/bin/bash

# Estrutura principal
mkdir -p cmd/api
mkdir -p internal/domain/{entity,valueobject,port}
mkdir -p internal/usecase
mkdir -p internal/adapter/{http/{handler,middleware},postgres,redis,event}
mkdir -p pkg/{logger,httputil}
mkdir -p migrations
mkdir -p docs/{adr,swagger}
mkdir -p scripts
mkdir -p .github/workflows
mkdir -p docker/postgres

# Arquivos base
touch cmd/api/main.go
touch docker/postgres/init.sql
touch scripts/.gitkeep
touch .env.example
touch Makefile
touch docker-compose.yml
touch Dockerfile

# Confirmar estrutura
find . -not -path './.git/*' -not -name '.gitkeep' | sort