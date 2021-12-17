# Envoy Resource Updater

The resource updater copies the envoy.yaml, cluster.yaml and listener.yaml from a configmap to a target folder. It does that by copying the files to an intermediate 
file in the target-folder and then renames them to trigger the move-event that envoy needs for reloading the config.

Configuration:

| Env-Var  | Description                                                     | Default       |
|----------|-----------------------------------------------------------------|---------------|
| PROXY_CONFIG_FOLDER | Folder to write the listener, envoy, and cluster yaml files to. | /proxy-config |
| CONFIG_MAP_FOLDER | The folder where the configmap is mounted to. | /configmap-folder |
| RUN_AS_INIT | Should the container run as an init-container, e.g. also copy the envoy.yaml? | false |