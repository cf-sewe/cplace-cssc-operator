# cplace Self-Service Cloud: Operator

The cSSC Operator is responsible for managing cplace instances:

- Uses GIT to obtain instance configurations.
- Provides a HTTP API:
  - Implements a webhook to retrieve notifications from GitHub about instance configuration changes.
  - Implements the cSSC Controller interface for provide environment and instance status as well as possibility to trigger actions.

## Design

The Operator is stateless - it does not require a database.
cplace instance configuration is provided by GIT and instance status is determined on-the-fly from the running system.

### Dependencies

We use the following dependencies:

- [gin-gonic/gin](https://github.com/gin-gonic/gin): Web framework
- [rookie-ninja/rk-boot](https://github.com/rookie-ninja/rk-boot): Bootstrapper
- [go-git/go-git](https://github.com/go-git/go-git): GIT interface

For classic stack:
- [cloudflare/cloudflare-go](https://github.com/cloudflare/cloudflare-go): DNS management
- [Docker SDK](https://docs.docker.com/engine/api/sdk/)

## Stacks

### Classic

This stack uses Hetzner dedicated machines.
Multiple servers form a cSSC environment.
There is no technical cluster - servers are operating completely independent.
This stack does not support high availability or failover and a load balancer is not used.

cplace is running as Docker containers and the data is stored on Docker volumes backed by ZFS.
The cplace container will be built on-the-fly from the software downloaded from Central.

Responsibilities:

- Instance Management
  - cplace Update
  - DNS Management (creates Cloudflare DNS entries for each instance)
  - Supports Snapshot / Restore
  - Supports Admin Scripts
- Environment Management
  - Capacity of each server and environment.

### Nomad

TBD

### Kubernetes

TBD