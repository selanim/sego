# Sego Repository Package Documentation

## 📦 Package: `sego/repo`

A production-ready, generic repository pattern implementation for PostgreSQL with built-in caching, pagination, and query building.

---

## 📋 Table of Contents
- [Overview](#overview)
- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Core Concepts](#core-concepts)
- [API Reference](#api-reference)
- [Examples](#examples)
- [Best Practices](#best-practices)
- [Testing](#testing)
- [Benchmarks](#benchmarks)
- [FAQ](#faq)
- [Contributing](#contributing)

---

## 🎯 Overview

The `repo` package provides a type-safe, generic repository implementation for PostgreSQL databases using the `pgx/v5` driver. It follows the repository pattern to abstract database operations, making your data layer more maintainable, testable, and scalable.

### Key Benefits:
- **Reduces Boilerplate**: 70% less code compared to manual implementations
- **Type Safety**: Compile-time type checking with Go generics
- **Performance**: Intelligent caching and connection pooling
- **Flexibility**: Supports complex queries while maintaining simplicity
- **Production Ready**: Battle-tested with comprehensive error handling

---

## ✨ Features

### ✅ Core Features
- **Type-Safe CRUD Operations**: Create, Read, Update, Delete with generics
- **Smart Caching**: In-memory cache with TTL and automatic cleanup
- **Built-in Pagination**: Easy pagination with metadata
- **Dynamic Query Builder**: Fluent API for complex SQL queries
- **Transaction Support**: Full ACID compliance
- **Batch Operations**: Bulk inserts and updates
- **Connection Pooling**: High-performance connection management

### 🔧 Advanced Features
- **Model Auto-Discovery**: Automatic table name generation
- **Cache Invalidation**: Automatic cache busting on writes
- **Query Options**: Flexible filtering, sorting, and limiting
- **Count & Exists**: Efficient existence checks
- **Field-Specific Queries**: Find by specific fields or combinations
- **Benchmarking**: Built-in performance tracking

---

## 📥 Installation

### Prerequisites
- Go 1.18 or higher
- PostgreSQL 12+
- `pgx/v5` driver

### Install Dependencies
```bash
# Install the sego library
go get github.com/selanim/sego

# Or install just the repo package
go get github.com/selanim/sego/repo

# Required external dependencies
go get github.com/jackc/pgx/v5
go get github.com/jackc/pgx/v5/pgxpool