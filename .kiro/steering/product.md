---
inclusion: fileMatch
fileMatchPattern: ["**/README.md", "**/ARCHITECTURE.md", "**/docs/**/*.md"]
---

# AAA Service Business Context

## Service Purpose

Enterprise-grade Authentication, Authorization, and Accounting service providing JWT-based authentication, PostgreSQL RBAC with hierarchical roles, comprehensive audit logging, and multi-tenant organization management.

## Core Domain Concepts

### Authentication

- JWT tokens with refresh token support
- Multi-factor authentication capabilities
- User lifecycle management with profiles and contacts

### Authorization (RBAC)

- **Hierarchical roles** with organization and group scoping
- **Role inheritance** - users inherit roles from parent groups/organizations
- **Fine-grained permissions** with resource-level control
- **Real-time permission evaluation** using PostgreSQL

### Accounting & Audit

- Comprehensive audit logging for all operations
- Event tracking for compliance reporting
- Anonymous audit support for sensitive operations

### Multi-tenancy

- **Organizations** - top-level tenant isolation
- **Groups** - sub-organization units with role inheritance
- **Hierarchical structure** - organizations contain groups, groups contain users

## Key Business Rules

### Role Inheritance

- Users inherit roles from their groups
- Groups inherit roles from parent groups
- Organizations can assign roles to groups
- Role inheritance flows down the hierarchy: Organization → Group → User

### Permission Evaluation

- Permissions are evaluated in real-time from PostgreSQL
- Users get effective permissions from all inherited roles
- Permission checks consider organization and group context

### Audit Requirements

- All user actions must be auditable
- Sensitive operations support anonymous audit trails
- Audit logs include context (organization, group, user, action, resource)

## Target Use Cases

- Enterprise applications with complex permission hierarchies
- Multi-tenant SaaS platforms requiring organization isolation
- Microservices needing centralized authentication and authorization
- Systems requiring detailed compliance and audit reporting
- Applications with hierarchical organizational structures

## Integration Patterns

- **HTTP REST API** for web applications
- **gRPC** for high-performance service-to-service communication
- **JWT tokens** for stateless authentication
- **Public client packages** for external service integration
