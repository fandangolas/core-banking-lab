# Core Banking Lab

[![Go Version](https://img.shields.io/badge/Go-1.20-blue)]() [![License: MIT](https://img.shields.io/badge/License-MIT-green)]()

## Overview

Core Banking Lab is an engineering sandbox inspired by real-world core banking systems. It explores safe concurrent operations, observability, authentication, CI/CD, and Kubernetes orchestration—simulating high-reliability financial infrastructure. :contentReference[oaicite:0]{index=0}

## Motivation

Traditional banking APIs often hide the complexity of concurrency control, infrastructure orchestration, and observability. With Core Banking Lab, you get a hands-on environment to experiment with:

- **Concurrency**: safe, high-throughput operations across multiple accounts  
- **Observability**: real-time metrics, logs, and dashboards  
- **Infrastructure**: containerization, Kubernetes, and CI/CD pipelines  

## Current Status

| Phase                                     | Status      |
|-------------------------------------------|-------------|
| 1. Architecture & Project Structure       | 🔲 Planned  |
| 2. Advanced Concurrency                   | 🔲 Planned  |
| 3. Testing                                | 🔲 Planned  |
| 4. Real-Time Simulation                   | 🔲 Planned  |
| 5. Observability                          | 🔲 Planned  |
| 6. Infrastructure & Deployment            | 🔲 Planned  |
| 7. CI/CD Automation                       | 🔲 Planned  |
| 8. Optional Features (JWT, scheduler, CLI)| 🔲 Planned  |
| 9. Portfolio Presentation (README, GIFs)  | 🔲 Planned  |

### Highlights to Date

- Layered architecture with `handler`, `service`, `repository`, and `domain` packages  
- Defined repository interfaces for both in-memory and PostgreSQL backends  
- Implemented per-account `sync.Mutex` locks with ordered acquisition to avoid races and deadlocks  
- Established initial unit & integration tests using `go test -race`  

### Critical Gaps

- **PostgreSQL Adapter** still needs a concrete implementation and validation tests  
- **Swagger/OpenAPI** documentation for endpoints is missing  
- No **observability** (Prometheus/Grafana) or **CI/CD** pipelines configured  
- Lacks **concurrency diagrams** and **benchmark results** to differentiate from a basic CRUD demo :contentReference[oaicite:1]{index=1}

## Roadmap & Next Steps

> At this early stage, we have the basic package structure in place and a handful of initial tests. The next focus is to review and expand these foundations:

1. **Core Architecture Review**  
   - Audit and refine package boundaries (`handler`, `service`, `repository`, `domain`)  
   - Validate and refactor existing unit tests for coverage and clarity  

2. **PostgreSQL Integration**  
   - Implement the `repository/postgres` adapter with basic CRUD operations  
   - Add containerized integration tests (Docker / Testcontainers) to verify data persistence  

3. **Comprehensive Testing & Benchmarks**  
   - Expand unit and integration test coverage, including edge cases and concurrency scenarios  
   - Introduce simple load benchmarks to measure throughput and latency under stress  

4. **API Specification**  
   - Draft an OpenAPI/Swagger definition for all core endpoints  
   - Generate and validate client stubs against the spec  

5. **Observability Foundations**  
   - Instrument key paths with Prometheus metrics (transactions/sec, error rates)  
   - Prototype Grafana dashboards to visualize system health  

6. **CI/CD Pipeline Setup**  
   - Create a GitHub Actions workflow for build, lint, test and container builds  
   - Automate deployment artifacts to a container registry  

7. **Documentation & Differentiation**  
   - Produce concurrency flow diagrams and benchmark reports  
   - Capture lessons learned, architectural trade-offs and next experiment ideas  


## Getting Started

Make sure you have Go (1.20 or later) installed on your machine.

```bash
git clone https://github.com/fandangolas/core-banking-lab.git
cd core-banking-lab
go run src/main.go
