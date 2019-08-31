# Development
This document explains you how to get started with developing harbor-sync. It shows you how to install the prerequisites and how to build, test and run the controller.

## Get the code

```bash
$ git clone https://github.com/moolen/harbor-sync.git ~/dev/harbor-sync
$ cd ~/dev/harbor-sync
```

## Installing the test environment

### Prerequisites:

* [Vagrant](https://www.vagrantup.com/docs/installation/) must be installed
* [Minikube](https://github.com/kubernetes/minikube/releases) must be installed
* [Kubebuilder](https://book.kubebuilder.io/quick-start.html#installation) must be installed

Use the provided `Vagrantfile` to spin up a harbor instance.

```sh
$ vagrant up
```

Right now you need to click your way through harbor to create the projects for testing.
Once the installation is done harbor tells you the ip address for this installation (e.g. `http://172.28.128.XXX.xip.io.`).

Tell the manager to access this deployment using environment variables:

```sh
$ export HARBOR_API_ENDPOINT=http://172.28.128.XXX.xip.io.
$ export HARBOR_USERNAME="admin"
$ export HARBOR_PASSWORD="Harbor12345"
```

Next, deploy the CRD and run the controller:
```
$ make generate # gen crds & manifests
$ make install # install crds
$ make run
```

## Developing

Now you're set to do your changes.
Please keep in mind:

* if you add a feature, please add documentation about the usage and write tests that cover at least the happy path


### Commit Messages

This projects follows the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0-beta.2/#summary) specification.

### Reconciliation loop

This is pretty straight-forward:

* find harbor projects that match the configured regular expression
  * reconcile robot accounts: i.e. (re-)create them if they do not exist, are disabled, expired or we do not manage the token
* find namespaces using a `mapping` config
  * for each namespace: create a secret with type `dockerconfigjson` with the specified name.

The reconciliation loop is triggered from essentially three sources:
* Control Plane: whenever a SyncConfig is created/updated/deleted
* Harbor Polling: whenever the state in harbor changes (project or robota account is created, updated, deleted)
* time-based using the configured `force-sync-interval`: forces reconciliation in a fixed interval to cover cases like namespace creation or robot account expiration

### Architecture

![Architecture]({{< baseurl >}}/harbor-sync-dev.jpg)
