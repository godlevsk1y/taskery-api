# Taskery REST API

REST API для **Taskery** — системы управления задачами и списками дел с синхронизацией, предназначенной для использования через CLI-клиент.

## Назначение

Этот проект создан для практики:
- проектирования backend-приложения с использованием принципов DDD (Domain-Driven Design)
- разделения доменной логики, прикладного и инфраструктурного слоёв
- построения чистого и поддерживаемого REST API

## Архитектура

Приложение структурировано по слоям:

- **Доменный слой (Domain layer)**  
  Основная бизнес-логика: сущности, value objects, доменные правила

- **Прикладной слой (Application layer)**  
  Сценарии использования и сервисы, которые координируют доменную логику

- **Инфраструктурный слой (Infrastructure layer)**  
  Доступ к данным и внешние интеграции (например, провайдер JWT)

- **Транспортный слой (Transport layer)**  
  REST API (HTTP-обработчики, DTO)

## Технологии

- Go
- REST API
- PostgreSQL
- JWT-авторизация
---
# Taskery REST API

REST API for **Taskery** — a task and todo management system with synchronization
designed to be used by a CLI client.

## Purpose

This project is created to practice:
- Designing a backend application using DDD principles
- Separating domain logic from application and infrastructure layers
- Building a clean and maintainable REST API

## Architecture

The application is structured into layers:

- *Domain layer*  
  Core business logic: entities, value objects, domain rules

- *Application layer*  
  Use cases and services that orchestrate domain logic

- *Infrastructure layer*  
  Data access and external integrations (e.g. JWT provider)

- *Transport layer*  
  REST API (HTTP handlers, DTOs)

## Technologies

- Go
- REST API
- PostgreSQL
- JWT Authorization
