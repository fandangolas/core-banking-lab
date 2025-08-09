# Architecture Decisions

## Diplomat Architecture

We adopt a **diplomat-architecture**, our variant of the classic Ports and Adapters pattern. Core domain rules stay isolated from external systems, while adapters act as diplomats translating between the domain and the outside world.

This separation keeps modules well organized, improves clarity of responsibilities, and makes it straightforward to plug in new technologies without rewriting business logic. By decoupling components, the system can scale and evolve as integrations grow in number and complexity.

## Database Strategy

For the moment, all data lives in an in-memory store to favor simplicity and fast iterations. In the future, the project will move to **PostgreSQL** to gain durability, richer querying, and stronger concurrency guarantees.

## Kubernetes for Real-World Experiments

Even though this is just a lab, the services will run on **Kubernetes**. Container orchestration lets us simulate real-world scaling scenarios, explore service discovery and resilience, and prepare the codebase for production-grade deployments.

