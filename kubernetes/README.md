# Deployment on kubernetes

The example inside this folder provides two configurations:
- [test-service setup](test-service.yaml)
- [helm values for the endpoint-auth-service](values.yaml)

The test-service setup consists of 2 echo-servers and a service to one of them:
- deployment "injection": the echo-server that should get the proxy injected
- deployment "additional": the echo-server that should serve as a target for receiving the request
- service "additional": service-interface to the "additional" deployment

Apply the setup with ```kubectl apply -f test-service.yaml```. Be aware that the setup gets deployed into the namespace ```proxy-test``` to comply with the certificate
in the helm-values. If the namespace needs to be different, update the values.yaml according to the [chart-documentation](https://github.com/FIWARE/helm-charts/tree/main/charts/endpoint-auth-service#sidecar-injection).

The setup should look similar to:
```shell
    kubectl get pods -n proxy-test
    
    NAME                                                         READY   STATUS    RESTARTS   AGE
    additional-6c6897ff88-mdt55                                  1/1     Running   0          10s
    injection-545558c46-hm46n                                    1/1     Running   0          10s
```


To deploy the endpoint-auth-service, use:

```shell
    helm repo add fiware https://fiware.github.io/helm-charts
    helm repo update
    helm install test-proxy fiware/endpoint-auth-service -n proxy-test -f values.yaml
```

After the service is deployed, the "injection"-pod needs to be recreated in order to trigger the actual injection:

```shell
    kubectl delete pod <injection-pod-id> -n proxy-test
```

A couple of seconds later, it should look like:

```shell
    kubectl get pods -n proxy-test
    
    NAME                                                         READY   STATUS    RESTARTS   AGE
    additional-6c6897ff88-mdt55                                  1/1     Running   0          10m
    injection-545558c46-hm46n                                    3/3     Running   1          10m
```
You can see, that the injection pod now has 3 containers(echo-server, sidecar-proxy, resource-updater)

Access the services via proxy:
```shell
  kubectl proxy --port 8002
```

and configure the endpoint:
```shell

  curl -X 'POST' \
      'http://localhost:8002/api/v1/namespaces/proxy-test/services/test-proxy-cs-endpoint-auth-service:8080/proxy/endpoint' \
      -H 'accept: */*' \
      -H 'Content-Type: application/json' \
      -d '{
          "domain": "additional-service",
          "port": 80,
          "path": "/notification",
          "useHttps": false,
          "authType": "iShare",
          "authCredentials": {
            "iShareClientId": "string",
            "iShareIdpId": "string",
            "iShareIdpAddress": "https://ar.isharetest.net/connect/token",
            "requestGrantType": "client_credentials"
          }
        }'
```

and your credentials for ishare(they need to be valid with the configured satellite):
```shell
  curl -X 'POST' \
      'http://localhost:8002/api/v1/namespaces/proxy-test/services/ishare-auth:8080/proxy/credentials/iShareClientId' \
      -H 'accept: */*' \
      -H 'Content-Type: application/json' \
      -d '{
          "certificateChain": "string",
          "signingKey": "string"
        }'
```

To test everything, now exec into the injected echo-server:
```shell
    kubectl exec -it <injection-pod-id> -c echo bash -n proxy-test
    
    curl -X GET additional-service/notification
    
    Result:
      CLIENT VALUES: ...
      SERVER VALUES:
      server_version=nginx: 1.10.0 - lua: 10001
      HEADERS RECEIVED:
        accept=*/*
        authorization=<BASE64-encoded-token>
```
The ```authorization```-header responded by the echo-server was added by the proxy.