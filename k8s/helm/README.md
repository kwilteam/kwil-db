Helm charts for Kwil.

## charts dependencies

```
                     +--------------+
                     |              |
        +------------+      kgw     +-----------+
        |            |              |           |
        |            --------+------+           |
        |                                       |
        |                                       |
        |                                       |
        |                                       |
        v                                       v
+-------+------+                       +--------+------+
|              |                       |               |
|     kwild    |                       |     hasura    |
|     (grpc)   |                       |    (graphql)  |
+-------+------+                       +--------+------+        
        |                                       |
        |                                       |
        v                                       |
+-------+------+                                |
|              |                                |
|  PostgreSQL  |<- - - - - - - - - - - - - - - -+        
| (local/aws)  |
+--------------+
```

Every chart could be deployed alone with its own dependency, eg. if you deploy `kwild` or `hasura`, there will be a postgres instance.

When deploy `kgw` locally, only one postgres instance will be created and shared between `kwild` and `hasura`.

When deploy `kgw` with an existing postgres(local or cloud), you need to overwrite default values, take a look at `kgw/staging-values.yaml` for reference.  

## initial setup

```
# in k8s/helm directory

# add bitnami repo
helm repo add bitnami https://charts.bitnami.com/bitnami

# update chart dependencies
helm dep update kwild
helm dep update hasura
helm dep update kgw
```

## local dev

To deploy kwil to local k8s cluster(assume using docker-desktop k8s), try:
```
# a local 'kwild:latest' and 'kwil-gateway:latest' image is needed
helm install kgw kgw -f kgw/def-values.yaml
```

Modify `kgw/dev-values.yaml` to overwrite default values for easier local development.

## development

Any sub-chart changes, need to call `helm dep update PARENT-CHART`.