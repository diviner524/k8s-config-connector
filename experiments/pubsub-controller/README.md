This is a simple demo to show how application team can create a composite resource.

* First install the CRD:

```shell
make install
```

* Then create a custom resource:
```shell
kubectl apply config/samples/infra_v1alpha1_pubsub.yaml
```

* Then run the controller locally to see how resources are provisioned/created
```
make run
```

The resources are provisioned based on the template defined in file.yaml
```
