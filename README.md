# cplace Self-Service Cloud: Operator

The cSSC Operator is responsible for managing cplace instances:

- Uses GIT to obtain instance configurations.
- Provides an HTTP API:
  - Implements a webhook to retrieve notifications from GitHub about instance configuration changes.
  - Implements the cSSC Controller interface for providing environment and instance status as well as the possibility to trigger actions.

- [cplace Self-Service Cloud: Operator](#cplace-self-service-cloud-operator)
  - [Design](#design)
    - [Dependencies](#dependencies)
  - [Stacks](#stacks)
    - [Classic](#classic)
    - [Nomad](#nomad)
    - [Kubernetes](#kubernetes)

## Design

The Operator is stateless - it does not require a database.
cplace instance configuration is provided by GIT and instance status is determined on-the-fly from the running system.

### Dependencies

We use the following dependencies:

- [gin-gonic/gin](https://github.com/gin-gonic/gin): Web framework
- [rookie-ninja/rk-boot](https://github.com/rookie-ninja/rk-boot): Bootstrapper
- [go-git/go-git](https://github.com/go-git/go-git): GIT interface

For classic stack:
- [cloudflare/cloudflare-go](https://github.com/cloudflare/cloudflare-go): Cloudflare DNS management
- [Docker SDK](https://docs.docker.com/engine/api/sdk/): cplace container management

## Stacks

### Classic

This stack relies on Hetzner dedicated machines for running cplace instances cost-efficiently.
Multiple servers form a cSSC environment, but there is no technical cluster.
The servers are operating completely independently.
High availability, failover and use of load balancer are not supported.

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

## Example Flow

### cplace Instance is Deployed

The cSSC Controller deploys as cplace instance.
The instance belongs to a user/organization, however the Operator component is not aware of users or organizations.

#### User Inputs

There are several supported use cases how a user may deploy a cplace instance to the cSSC.
The user inputs will be stored by the Controller in GIT along with other instance specific configuration.

The Controller component is responsible to validate user inputs before submitting to GIT.
For example, the Controller would confirm by calling the `GET /dns` API that the user selected instance domain name is available.

Demo case:

- User in cFactory or SSE organization selects environment "Demo", instance template
- Instance name is automatically generated: `demo-o3t8.cplace.cloud`
- Instance sizing is automatically configured (2 core, 3 GB heap, 25 GB disk) by selected template
- Release can be selected, but integration repo is automatically chosen (demo build)
- Instance lifetime: selection limited by instance template to: 30, 60, 90 days.
- Custom config:
  - Solution template selection

Dev case:

- User in cFactory or SSE Org selects environment "Test", instance template:
- Instance name: User can enter custom identifier, `<custom>.cf.test.cplace.cloud`
- Instance sizing is configured by selected template
- Build identifier from central can be selected (later Repo, Release + Build)
- Instance lifetime: selection limited by instance template to: 6mo, 1yr, 2yr.
- Custom config:
  cplace properties

#### Preparation

- checkDns: Check that the desired instance DNS is valid and available.
- checkCapacity: Check that the cluster has enough resources available for the new instance.
- getSupportedCplaceReleases: Returns list of supported cplace versions

GET environment -> return env info, including base domain (e.g. `test.cplace.cloud`)
GET instance

instance/prepare
instance/deploy
instance/status

isDnsAvailable test.cplace.cloud
checkCapacity  ram: 4G, disk: 50GB, cpu 4

## cplace Release Management

We want to support a user-friendly way to select the intended cplace build:

- An organization has accessible cplace integration repositories, for example, customer specific integration repos.
- A user is part of an organization and has implicit access to the cplace integration repos.
- This mapping is contained in Controller domain.
- The user can select the desired cplace release from the Controller UI:
  - User selects a cplace repository
  - Controller uses info from Operator API and own definition which cplace release versions are supported (e.g. 22.4, 23.4).
    The Operator needs to distinguish cplace release due to infrastructure differences (e.g. Java, Elasticsearch version).
    The Controller distinguishes cplace release versions due to configuration differences.
  - User selects the desired cplace release version.
  - Controller populates list of builds

The cplace Container is prepared like this (at least on Classic Stack):

- A base container with the required cplace dependencies is provided.
  Different cplace releases might use a different base container.
- When the user selected a valid cplace build, the build's `software.zip` is downloaded from Central.
  The software.zip is extracted to a volume mount.

## Instance Configuration

All instances of an environment are managed by a GIT repository.
The instances are maintained in the following structure:

cplace-cssc-env-test (cloned GIT repo):
  /instances:
    sewe.cf.test.cplace.cloud/config.yaml

## APIs

### Environment API

> `GET /environment`

Returns environment information:

- configuration:
  - name: Environment name
  - description: Environment description
  - type: Type of backend stack (classic, nomad, kubernetes...)
  - baseDomain: Base domain of the environment, e.g. `test.cplace.cloud`
  - git:
    - repo: repository of the environment specific instance configuration
    - branch: GIT branch
- status:
  - instanceCount
  - nodeCount
  - capacity
    - memory
    - disk
    - cpu
- cplace:
  - supportedReleases: List of supported cplace releases, e.g. [23.1, 23.2, 23.3]

Other operations are not supported for the /environment API.

### Instances API

> `GET /instances`

Returns information of all instances or instances matching the filters:

- list of instances:
  - name: Instance name (unique identifier), e.g. `sewe.cf.test.cplace.cloud`
  - owner:
    - name: e.g. sebastian.weitzel@cplace.com
    - organization: e.g. cFactory
  - status: running, stopped, failed
  - capacity:
    - memory
    - diskTenant
    - diskDatabase
    - cpu
  - cplace:
    - repository:
    - release
    - branch
  - events
    - timestamp, created
    - timestamp, cplace update
    - timestamp, admin script
    - timestamp, configuration change

Params:
- name: filter by instance name (domain)
- release: filter by installed cplace release

> `GET /instances/{instanceId}`

Returns information for a specific instance (see above).

> `GET /instances/{instanceId}/log`

Returns the logs of the specified instance.
Intended to be used by Controller to stream logs real-time (see Spring Boot Admin).

TBC: log streaming? Console log? separated by log file? Log source Grafana Loki?

> `GET /instances/{instanceId}/metrics`

Returns the basic metrics of the specified instance.
Intended to be used by Controller to display instance status.
This API just returns the latest value.

Advanced metrics will be displayed in Grafana.
This API just provides basic metrics:

- Healthy (0/1)
- Uptime
- CPU
- Heap memory usage
- Used storage (per DB, tenant files)

> `GET /instances/{instanceId}/events`

Gets the events of the specified instance.
Any action on the instance will generate an event.
The Controller will display events to the user.

Events are stored in a meta JSON in a file stored per instance, for example under `/instances/sewe.cf.test.cplace.cloud/events.json`.

Params:
limit=XX : only display XX events (default 10)
offset=XX : skip XX events (default 0)

> `POST /instances/{instanceId}/restart`

Restarts the specified instance.

> `POST /instances/{instanceId}/update`

Performs a cplace update of the specified instance.
The step requires that the new release is specified in the updated instance configuration in GIT.
If configuration is changed, any change is also automatically applied.

#### Snapshots API

The snapshot location is defined in the Operator config.
Snapshot information is stored in "meta" files directly in that location.

Snapshots for the instance sewe.cf.test.cplace.cloud are stored for example under `/instances/sewe.cf.test.cplace.cloud/snapshots/<snapshot>.zip`.
A meta JSON file is stored under `/instances/sewe.cf.test.cplace.cloud/snapshots.json`.

> `GET /instances/{instanceId}/snapshots`

Lists all snapshots of the specified cplace instance.

- List of snapshots:
  - name: Snapshot name
  - created: Snapshot creation timestamp
  - release
  - location: Location, where the snapshot is stored
  - size

Params:

- release: cplace release the snapshot was created with.
- limit: retrieve only specified number of snapshots (default 10)
- offset: 0

> `POST /instances/{instanceId}/snapshots`

Performs a cplace snapshot using Tenant export functionality.
Such a snapshot may only be restored to the same cplace release version.

This API does not limit number of Snapshots created, but the Controller should.
Snapshots will not currently be automatically removed (later they should).

> `GET /instances/{instanceId}/snapshots/{snapshotId}`

Retrieves information for the specified snapshot and instance (see above).

> `DELETE /instances/{instanceId}/snapshots/{snapshotId}`

Deletes the specified snapshot.
Note that snapshots may also be removed from the file system.
Therefore, the Operator should scan snapshots regularly for consistency.

## Unsorted Ideas

- Test strategy could involve a mocked stack that installs a very simple HTTP application.
- Instance prolongation should be possible by user from the Controller UI later (auto prolong maybe even).
  Operator is responsible to clean expired instances and backups (30 days after instance removal).
- Custom Domain (CNAME) will only be supported later, for production use cases.
- Instances can be matched against cost center by the Controller.
- One environment can be used by multiple organizations (e.g. Dev, Sales, Partner).
  The Controller defines which org has access to which environments (TBC).
- Operator should provide metrics for the environment and instances (e.g. instance CPU usage).
  - The metrics can be consumed by Controller and displayed to the user.
