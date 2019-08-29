# Developing harbor-sync
This document explains you how to get started with developing harbor-sync. It shows you how to install the prerequisites and how to build, test and run the controller.

## Get the code

```bash
$ git clone https://github.com/moolen/harbor-sync.git ~/dev/harbor-sync
$ cd ~/dev/harbor-sync
```

## Installing the test environment

>**Prequisites**: Vagrant must be installed.
See [docs/installation](https://www.vagrantup.com/docs/installation/) for installation instructions.

>**Prequisites**: Minikube must be installed.
See [releases](https://github.com/kubernetes/minikube/releases) for installation instructions.

>**Prequisites**: kubebuilder must be installed.
See [docs/installation](https://book.kubebuilder.io/quick-start.html#installation) for installation instructions.

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

### Testing

If you have installed kubebuilder you can run the tests directly on your host machine using `make test`. If not, run the tests in a container:

```bash
# run testsuite in docker container (so you don't need to install kubebuilder)
$ make test-docker
```
