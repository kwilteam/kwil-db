Helm charts for Kwil.

## setup

### enable docker-desktop kubernetes

Enable Kubernetes in Docker Desktop UI.

### helm setup

```
# macos
brew install helm

# add bitnami repo(need postgresql)
helm repo add bitnami https://charts.bitnami.com/bitnami

# update chart dependencies
helm dependency update deployments/helm/hasura
helm dependency update deployments/helm/kwild
helm dependency update deployments/helm/kwil
```

## local dev

To deploy kwil(full deployment) to local k8s cluster(assume using docker-desktop k8s), try:
```
## to ease the local development, every task will update chart dependencies first.

# build all required images and install
task k8s:kwil

# install using existing images
task k8s:kwil:install

# uninstall
task k8s:kwil:uninstall

# When all pods from `kubectl get all` are running and ready, everything is ready to go.
```

Modify `kwil/dev-values.yaml` to overwrite default values for easier local development.

## chart development

⚠️ Any sub-chart changes, need to call `helm dep update PARENT-CHART`.

## charts dependencies

```
              +--------------+
              |              |
              +      kwil    +
              |      (gw)    |
              +-------+------+
                      |
                      |
                      |
                      v
              +-------+------+
              |              |
              |     kwild    +
              |     (grpc)   |
              +-------+------+
                      |
                      |
                      |
                      v
              +-------+------+
              |              |
              |     hasura   +
              |              |
              +-------+------+
                      |
                      |
                      |
                      v
              +-------+------+
              |              |
              |  PostgreSQL  |
              | (local/aws)  |
              +--------------+
```

Every chart could be deployed alone with its own dependency, eg. if you deploy `kwild` or `hasura`, there will be a postgres instance.

When deploy `kwil` with an existing postgres(local or cloud), you need to overwrite default values, take a look at `kwil/staging-values.yaml` for reference.
