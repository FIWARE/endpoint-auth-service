# Configmap updater

The configmap-updater is a small tool to publish the cluster.yaml and listener.yaml to a [config-map](https://kubernetes.io/docs/concepts/configuration/configmap/).
It holds a file-watcher to the mounted folder and executes the update on change. 

Configuration:

| Env-Var  | Description                                              | Default |
|----------|----------------------------------------------------------|---------|
| PROXY_CONFIG_FOLDER | Folder to load the listener and cluster yaml files from. |    /proxy-config     |
| PROXY_CONFIG_MAP | The configmap to be updated.                             |    envoy-config     |
| PROXY_CONFIG_MAP_NAMESPACE | Namespace the configmap is located in.                   |    envoy     |