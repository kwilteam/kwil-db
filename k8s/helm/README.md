Helm charts for Kwil.

## charts dependencies

```
                     +--------------+
                     |              |
        +------------+     kwil     +-----------+
        |            |              |           |
        |            --------+------+           |
        |                                       |
        |                                       |
        |                                       |
        |                                       |
        v                                       v
+-------+------+                       +--------+------+
|              |                       |               |
|  PostgreSQL  | <- - - - - - - - - - -|     Hasura    |
|    (local)   |                       |               |
+--------------+                       +---------------+
```

## initial setup

```
helm repo add bitnami https://charts.bitnami.com/bitnami

# update hasura chart dependency
cd kwil/charts/hasura
helm dependency update

# update kwil chart dependency
cd -
helm dependency update
```

## local dev

To deploy kwil to local k8s cluster(assume using docker-desktop k8s), try:
```
# a local 'kwild:latest' image is needed
helm install kwil kwil/ -f kwil/dev-values.yaml
```

Modify `kwil/dev-values.yaml` to overwrite default values for easier local development.