# cplace Self-Service Cloud: Operator

- [cplace Self-Service Cloud: Operator](#cplace-self-service-cloud-operator)
  - [Introduction](#introduction)
  - [Design](#design)
    - [Data Persistency](#data-persistency)
    - [Dependencies](#dependencies)
  - [Stacks](#stacks)
    - [Docker Swarm](#docker-swarm)
      - [Implementation Information](#implementation-information)
      - [cplace Container Preparation](#cplace-container-preparation)
      - [Storage Layout](#storage-layout)
      - [Instance Quota](#instance-quota)
  - [cplace Release Management](#cplace-release-management)
  - [Instance Configuration](#instance-configuration)
  - [APIs](#apis)
    - [Environment API](#environment-api)
    - [Instances API](#instances-api)
      - [Snapshots API](#snapshots-api)
    - [Admin Scripts API](#admin-scripts-api)
    - [DNS API](#dns-api)
  - [Example Flow](#example-flow)
    - [cplace Instance is Deployed](#cplace-instance-is-deployed)
      - [User Inputs](#user-inputs)
  - [Unsorted Ideas](#unsorted-ideas)
  - [Links](#links)

## Introduction

This service is called Operator and is responsible for managing cplace instance deployments for a specific environment via GIT and a RESTful API.
One instance of the operator is responsible for one environment.
Docker Swarm is used as the backing cplace container orchestration system.

The operator configuration is loaded from the OS environment variables or a `.env` file.
One environment contains cplace instances, that are managed by the operator.
The cplace instances are defined in a GIT repository, with one YAML file per instance.

When the Operator starts, it initializes the connection to the environment using the connection data specified in the env configuration.
It also starts a background worker that checks the GIT repository for instance definitions (including their configuration) and applies them to the environment regularly.

Then it initializes the gin framework HTTP routes for environment and instance management.
Example actions that may be triggered via API:

- `GET /environment`: Returns environment information, including status
- `GET /instances`: Returns information of all instances or instances matching the filters.
  Example information: status, cplace version info, used capacity.
- `GET /instances/{instanceId}`: Returns information for a specific instance.
- `GET /instances/{instanceId}/log`: Returns the logs of the specified instance.
- `GET /instances/{instanceId}/metrics`: Returns the basic metrics of the specified instance.
  Example metrics: healthy, uptime, CPU usage, heap memory usage, used storage.
- `GET /instances/{instanceId}/events`: Returns the events of the specified instance.
- `POST /instances/{instanceId}/restart`: Restarts the specified instance.
- `GET /instances/{instanceId}/snapshots`: Lists all snapshots of the specified cplace instance.
- `POST /instances/{instanceId}/snapshots`: Creates a cplace snapshot using Tenant export functionality.
- `GET /instances/{instanceId}/snapshots/{snapshotId}`: Retrieves information for the specified snapshot and instance.
- `DELETE /instances/{instanceId}/snapshots/{snapshotId}`: Deletes the specified snapshot.
- `GET /instances/{instanceId}/admin-scripts`: Lists all admin scripts of the specified cplace instance that have been executed.
- `POST /instances/{instanceId}/admin-scripts`: Uploads a new admin script and executes it.
- `GET /instances/{instanceId}/admin-scripts/{adminScriptId}`: Retrieves information for the specified admin script and instance.

## Design

### Data Persistency

The Operator does not have a database in the traditional sense, but it stores certain instance information on storage:

1. Snapshot information:
  Information about the snapshots that exist for a certain instance is collected in a JSON file, e.g. `/instances/sewe.cf.test.cplace.cloud/snapshots.json`.

2. Events:
  Information about all activities performed by the Operator for a specific instance is stored in a JSON file, e.g. `/instances/sewe.cf.test.cplace.cloud/events.json`.

3. Admin Scripts:
  Information about all admin scripts that are executed for a specific instance is stored in a JSON file, e.g. `/instances/sewe.cf.test.cplace.cloud/admin-scripts.json`

4. Software releases:
   Software releases are downloaded in the deployment phase of a cplace instance.
   To allow efficient deployments, the releases are stored at a central location and hard-linked to the containers that require that build.

cplace instance *configuration* is provided by GIT and instance status is determined on-the-fly from the running system.

Scanning the directories & files is inefficient, especially for real-time operations.
For example finding instances that are owned by a specific user requires scanning all instance directories.
Therefore, we use a cache for the instance information.
The cache is updated with instance operations or regularly in the background.

### Dependencies

We use the following dependencies:

- [go-chi/chi](https://github.com/go-chi/chi): Minimalistic web framework
- [go-git/go-git](https://github.com/go-git/go-git): GIT interface

For classic stack:
- [cloudflare/cloudflare-go](https://github.com/cloudflare/cloudflare-go): Cloudflare DNS management
- [Docker SDK](https://docs.docker.com/engine/api/sdk/): cplace container management

## Stacks

### Docker Swarm

The Docker Swarm stack relies on Hetzner dedicated machines for running cplace instances cost-efficiently.
Docker Swarm provides an HTTP API endpoint for managing the service deployments.
By using the Docker Swarm provided overlay network,
we can also use a wildcard domain (e.g. *.test.cplace.cloud) and use a Hetzner load-balancer as cluster entry-point.

The current implementation is using MariaDB and Elasticsearch instances that are deployed on each dedicated server.
We do not use ES or DB clusters.
This means that a cplace instance has to be bound to a specific Swarm node.
Therefore, we do not require a cluster file system for cplace or the cplace databases.
The intended use of this stack is for non-commercial cplace instances running as single nodes.

cplace is running as Docker containers and the data is stored on Docker volumes backed by ZFS with a quota.
The cplace container is lightweight and does not contain the cplace build.
The build is downloaded from Central to a shared storage and then hard-linked and mounted to the container.

Stack features:

- Instance Management:
  - cplace deployments and updates
  - Snapshot and restore functionality
  - Admin script execution
- Cluster capacity management

#### Implementation Information

swarm configured in "simple" mode:
  - no cluster/replicated volumes
  - no DB cluster (one DB per node)
  - cplace instances are tied to one node
  - Advantages: network communication (*.test.cplace.cloud -> hcloud LB -> dedi server)
operator runs in swarm cluster
operator connects to swarm endpoint
to deploy, operator iterates through the nodes, and the first node with enough resources will be used to deploy

cplace Init procedure:

- select available node         -> < 1s
- create database+db user       -> < 1s
- create volume for cplace data -> < 1s
- download software if missing  -> < 5min
- hard-link/bind mount software -> < 1s
- start cplace container        -> ±1 min (basic init...)

operator should support fast instance deployments.
deployments can run in parallel.
instance actions should run in separate "threads"

#### cplace Container Preparation

- For each cplace release, a Dockerfile and configuration template is prepared.
  This means that we do not need individual container images.
- When deploying an instance, the required software.zip is downloaded from Central and stored on a shared storage.
  The software.zip is downloaded to a central location and unzipped.
- For each instance:
  - a /data/instances/<instance>/data & /properties folder is created
  - the correct software is hard-linked to the instance folder
  - the configuration is created from the template for the specific cplace release
  - the databases are created

#### Storage Layout

| **Path** | **Size** | **Description** |
| --- | --- | --- |
| /boot | 1GB | Boot partition |
| / | 100GB | Root partition |
| ceph* | 100GB | ceph storage for cplace instance configuration and releases |
| rpool | rest | ZFS pool for data |

We need a cluster file system to make cplace instance configuration and releases accessible from each node.
We want to use a fraction of each node disk and add it to one large pool.
We can check ceph or Piraeus for this task.

Alternatively we can avoid using a shared storage.
The instance configs can be stored in Swarm configs.
The releases can be stored only on each node where its required.

#### Instance Quota

- Each cplace instance has a quota for tenant files and database size.
- tenant files are located on a ZFS volume (/data/instances/<instance>/data/tenants).
- database for the instance is located on a ZFS volume (/data/instances/<instance>/db).

## cplace Release Management

We want to support a user-friendly way to select the intended cplace build:

- An organization has accessible cplace integration repositories, for example, customer-specific integration repositories.
- A user is part of an organization and has implicit access to the cplace integration repositories.
- This mapping is contained in the Controller domain.
- The user can select the desired cplace release from the Controller UI:
  - User selects a cplace repository
  - Controller uses info from Operator API and own definition which cplace release versions are supported (e.g. 22.4, 23.4).
    The Operator needs to distinguish cplace release due to infrastructure differences (e.g. Java, Elasticsearch version).
    The Controller distinguishes cplace release versions due to configuration differences.
  - User selects the desired cplace release version.
  - Controller populates the list of builds

The cplace Container is prepared like this (at least on Classic Stack):

- A base container with the required cplace dependencies is provided.
  Different cplace releases might use a different base container.
- When the user selected a valid cplace build, the build's `software.zip` is downloaded from Central.
  The software.zip is extracted to a volume mount.

## Instance Configuration

All instances of an environment are managed by a GIT repository.
The instances are maintained in the following structure:

> cplace-cssc-env-test (cloned GIT repo):
>   /instances:
>     sewe.cf.test.cplace.cloud/config.yaml

TBD We could also have one repo for all environments and would then have the following structure:

> cplace-cssc-environments (cloned GIT repo):
>   /environments/test/instances:
>     sewe.cf.test.cplace.cloud/config.yaml

In the beginning we use one repository for all environments.
For production use (once introduced) we'd prefer the one-GIT-per-environment approach for compliance reasons.

Instance configuration (config.yaml):

```yaml
---
# TBC: maybe without base domain (make it easier to change it)
name: "sewe"
domain: "sewe.cf.test.cplace.cloud"
expiry: "2 days"
owner:
  user: "sebastian.weitzel@cplace.com"
  organization: "collaboration Factory"
  organization_short: "cf"
cplace:
  # always required
  release: "23.2"
  # specify cplace release by repo/branch
  repository: "cplace-ga-products"
  branch: "release/23.2"
  # or by repo/tag
  #tag: "23.2.0"
  # or by build id from central
  central_build_id: "e4gufovg2nf9hgvbgz8egc4gk"
  sizing:
    heap: "3072M"
    cpu: "2.0"
    disk: "20GB"
  configuration:
    # custom application.properties
    application: |
      logging.level.org.elasticsearch=DEBUG
    # custom JVM options
    jvm_options:
      - "-Dxxx"
  post-install:
    # Execute admin-scripts after initial cplace startup in the given order
    # These scripts must be uploaded to the GIT as well
    admin-scripts:
      - ImportTenantAction.java
      - CreateDemoUsers.java
```

## APIs

The following is a WIP documentation of the planned API.
The implemented API will be documented using Swagger/OpenAPI.

### Environment API

> `GET /environment`

Returns environment information:

- configuration:
  - name: Environment name
  - description: Environment description
  - type: Type of backend stack (Classic, Nomad, Kubernetes...)
  - baseDomain: Base domain of the environment, e.g. `test.cplace.cloud`
  - git:
    - repo: repository of the environment-specific instance configuration
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

When an instance is currently being deployed, its status is `deployment_<step>`.
This can be used by Controller for tracking the instance startup progress (and issues).

Deployment Steps:

- dns: DNS creation
- prepare_instance: Creation of basic instance structure, including file systems etc.
- build_container: Preparing the cplace container.
  Actually this step for Classic cloud only downloads and extracts the software.zip; the container itself is generic.
- start_container: Initial startup of the cplace container.
- adminscripts_<count>: Execution of initial admin scripts.

> `GET /instances`

Returns information of all instances or instances matching the filters:

- list of instances:
  - name: Instance name (unique identifier), e.g. `sewe.cf.test.cplace.cloud`
  - owner:
    - name: e.g. sebastian.weitzel@cplace.com
    - organization: e.g. "collaboration Factory"
    - organization_short: e.g. "cf"
  - status: running, stopped, crashed, crash_loop, deployment_<step> (deployment of the instance in progress)
  - status_details: Optional extra information
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

TBC: log streaming? Console log? Separated by log file? Log source Grafana Loki?

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
Any user or operator action on the instance will generate an event.
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
If the configuration is changed, any change is also automatically applied.

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
- limit: retrieve only a specified number of snapshots (default 10)
- offset: 0

> `POST /instances/{instanceId}/snapshots`

Performs a cplace snapshot using Tenant export functionality.
Such a snapshot may only be restored to the same cplace release version.

This API does not limit the number of Snapshots created, but the Controller should.
Snapshots will not currently be automatically removed (later they should).

> `GET /instances/{instanceId}/snapshots/{snapshotId}`

Retrieves information for the specified snapshot and instance (see above).

> `DELETE /instances/{instanceId}/snapshots/{snapshotId}`

Deletes the specified snapshot.
Note that snapshots may also be removed from the file system.
Therefore, the Operator should scan snapshots regularly for consistency.

### Admin Scripts API

The admin script location is defined in the Operator config.
Additional information is stored in "meta" files directly in that location.

Admin scripts for the instance sewe.cf.test.cplace.cloud are stored for example under `/instances/sewe.cf.test.cplace.cloud/admin-scripts/CreateTenantAction_arfia.java`.
A meta JSON file is stored under `/instances/sewe.cf.test.cplace.cloud/admin-scripts.json`.

Script execution will be monitored by the Operator.
The status of stale running scripts will be updated to "aborted" or "crashed" depending on the Operator's knowledge of the cplace instance.

Admin scripts cannot be canceled.

> `GET /instances/{instanceId}/admin-scripts`

Lists all admin scripts of the specified cplace instance that have been executed.
The information is retrieved from the meta JSON file.

- List of admin-scripts:
  - name: Admin-Script name
  - createdAt: Creation timestamp
  - completedAt: Completion timestamp
  - status: success, failed, running, aborted (cplace regular restart), crashed (cplace crashed)

Params:
- status: filter by status
- limit: retrieve only a specified number of snapshots (default 10)
- offset: 0

> `POST /instances/{instanceId}/admin-scripts`

Uploads a new admin script and executes it.
This API will not wait for its completion.
When the Operator has successfully scheduled the script for execution,
the returned `adminScriptId` can be used to check for the script's status.

> `GET /instances/{instanceId}/admin-scripts/{adminScriptId}`

Retrieves information for the specified admin script and instance (see above).

### DNS API

> `GET /dns`

Returns the DNS information of the specified instance domain.
Should be used by Controller to determine if a user-selected instance name is available.

HTTP 404 when the DNS entry was not found.

HTTP 200 when it exists and returns its information:
- type: Type of DNS record (e.g. `A`)
- value: Target IP or name (e.g. `1.2.3.4`)

Creating DNS entries via API is currently not planned and is implicitly done by the deployment procedure.

## Example Flow

### cplace Instance is Deployed

The cSSC Controller deploys a cplace instance.
The instance configuration is stored on GIT.
The operator is informed by a webhook of a GIT change or checks it at regular intervals.
The instance belongs to a user/organization, however, the Operator component is not aware of users or organizations.

#### User Inputs

There are several supported use cases how a user may deploy a cplace instance to the cSSC.
The user inputs will be stored by the Controller in GIT along with other instance-specific configuration.

The Controller component is responsible to validate user inputs before submitting to GIT.
For example, the Controller would confirm by calling the `GET /dns` API that the user-selected instance domain name is available.

Demo case:

- User in cFactory or SSE organization selects environment "Demo" (instance template is automatically selected)
- Instance name is automatically generated: `o3t8.cf.demo.cplace.cloud`
- Instance sizing is automatically configured (2 core, 3 GB heap, 25 GB disk) by template
- Release can be selected, but integration repo is automatically chosen (demo build)
- Instance lifetime: selection limited by instance template to: 30, 60, 90 days.
- Custom config:
  - Solution template selection
- Required admin scripts are provided to Operator.
  For example, the `SolutionTemplateImportAction.java` admin script will perform all required steps to provision a Solution Template demo.
  For example, it downloads the correct demo tenant archive for the currently running cplace release.

Dev case:

- User in cFactory or SSE Org selects environment "Test", instance template:
- Instance name: User can enter custom identifier, `<custom>.cf.test.cplace.cloud`
- Instance sizing is configured by selected template
- Build identifier from central can be selected (later Repo, Release + Build)
- Instance lifetime: selection limited by instance template to: 6mo, 1yr, 2yr.
- Custom config:
  cplace properties

Admin scripts are an essential part how instances are provisioned in the cSSC.
The tenant export files required for demo provisioning should be downloaded from a public web server.
Storing them is out of scope of the Operator.

## Unsorted Ideas

- Test strategy could involve a mocked stack that installs a very simple HTTP application.
- Instance prolongation should be possible by user from the Controller UI later (auto prolong maybe even).
  Operator is responsible to clean expired instances and backups (30 days after instance removal).
- Custom Domain (CNAME) will only be supported later, for production use cases.
  The problematic part here is the Cloudflare integration / extra cost for CNAME's.
- Instances can be matched against cost center by the Controller.
- One environment can be used by multiple organizations (e.g. Dev, Sales, Partner).
  The Controller defines which org has access to which environments (TBC).
- Operator should provide metrics for the environment and instances (e.g. instance CPU usage).
  - The metrics can be consumed by Controller and displayed to the user.
- Instances are monitored by the Operator.
  When an instance is crashing (depending on the backend stack) the instance will automatically be restarted until the attempts are exceeded.
- Private container registry running in Swarm Stack.
  Images are regularly cleaned from it if too old and unused.

## Links

- Shared storage: https://linbit.com/blog/create-a-docker-swarm-with-volume-replication-using-linbit-sds/




Worker procedure:

- clone GIT
- validate GIT structure
- load git instance config
- iterate through instances from GIT and perform actions if required:
  - deploy
  - delete
  - config change / update
- the actions should be async and return fast to the iterating loop

Orphaned instances:
  - instance running in cluster but has no GIT definition
  - can happen if the GIT was manually edited.
    Note: When GIT structure is invalid, this should be prevented and not cause orphaned instances!
  - GetInstances() can return orphaned instances, filter supports orphaned flag
    controller can then show / cleanup orphaned instances (e.g. by admins)

Instance Locking:
  - If a change request is already processed for an instance, the worker loop should skip the instance until the next run.
    That means instance jobs need to be tracked for completion.

Keeping track of activities:
  - each Instance object Status indicates whats going on.
  - e.g. when the instance is being deployed, the status will contain the relevant information, including the phase of deployment

Robustness:
  - each action should be able to recover, in case it was aborted violently
  - e.g when the deployment is restarted, the procedure should only run steps if needed and perform the missing steps.
    basic idempotence.
