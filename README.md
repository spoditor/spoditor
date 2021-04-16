# Spoditor

Spoditor, (StatefulSet Pod Editor), is a Kubernetes dynamical admission controller to differentiate each individual Pod belonging to a StatefulSet.

[StatefulSet](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/) is designated to manage stateful workload on Kubernetes. However, developer often has to face the limit that all the Pods in a StatefulSet have to share the same PodSpec. A lot of stateful workload cluster actually requires slightly different specification, such as configuration, storage, etc, on one or more Pods.

Spoditor helps to lift this limitation of StatefulSet, by allowing developer to annotate the PodSpec template of a StatefulSet and apply extra configuration to Pod of different ordinal.

## Quick Example
```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: web
spec:
  selector:
    matchLabels:
      app: nginx # has to match .spec.template.metadata.labels
  serviceName: "nginx"
  replicas: 3 # by default is 1
  template:
    metadata:
      labels:
        app: nginx # has to match .spec.selector.matchLabels
      annotations:
        spoditor.io/mount-volume: |
          {
            "volumes": [
              {
                "name": "my-volume",
                "secret": {
                  "secretName": "my-secret"
                }
              }
            ],
            "containers": [
              {
                "name": "nginx",
                "volumeMounts": [
                  {
                    "name": "my-volume",
                    "mountPath": "/etc/secrets/my-volume"
                  }
                ]
              }
            ]
          }
    spec:
      terminationGracePeriodSeconds: 10
      containers:
      - name: nginx
        image: k8s.gcr.io/nginx-slim:0.8
        ports:
        - containerPort: 80
          name: web
        volumeMounts:
        - name: www
          mountPath: /usr/share/nginx/html
  volumeClaimTemplates:
  - metadata:
      name: www
    spec:
      accessModes: [ "ReadWriteOnce" ]
      resources:
        requests:
          storage: 1Gi
```
The annotation `spoditor.io/mount-volume` above will mount secret `my-secret-0` to container `nginx` in Pod `web-0`, secret `my-secret-1` to Pod `web-1`, secret `my-secret-2` to Pod `web-2`, so on and so forth.

This annotation takes a JSON object as its value, the schema of `volumes` and `containers` fields are the same as corresponding fields in [PodSpec](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#PodSpec), though only subset of the schemas are meaningful to this annotation, irrelevant fields are just simply ignored.

**Important**

Obviously, Spoditor does not create the _expanded_ secret, configmap, etc resources for each Pod. It is the client's responsibility to provide them. If an _expanded_ resource does not exist, the corresponding Pod will be pending for container creation.

## Annotation Qualifier

The example above illustrates mounting dedicated secret (potentially, other mountable resources) for each Pod matching the same ordinal.

However, in other scenario, developer may only want to argument a subset of Pods in a StatefulSet, for example, Pod 0 being the master node of a stateful workload cluster. Spoditor supports **qualifier** suffix for this purpose.

`spoditor.io/mount-volume[_{lower}-{upper}]`

| With qualifier suffix  | Applicable Pod ordinal |
| ------------- | ------------- |
| spoditor.io/mount-volume_0 | Only Pod 0  |
| spoditor.io/mount-volume_5-  | All Pod with ordinal >= 5 |
| spoditor.io/mount-volume_-5  | All Pod with ordinal <= 5 |
| spoditor.io/mount-volume_2-5  | All Pod with ordinal >= 2 AND <= 5 |

Multiple annotations with different qualifier suffix can be applied to the same StatefulSet. For example, we can use both `spoditor.io/mount-volume_0` and `spoditor.io/mount-volume_1-` to give Pod 0 a dedicated configuration while making all the other Pods share a same configuration.

## Editing Existing StatefulSet

Spoditor chooses to use annotations under the `.spec.template.metadata.annotations` field of a StatefulSet. This allows the reconciliation loop of the StatefulSet controller to kick in upon any update to any annotation, which means developer can argument running StatefulSet, and the underlying Pods will be recreated with dedicated configuration applied by Spoditor.

## Supported Annotations
### mount-volume
This annotation allows mounting different `secret` or `configmap` as volume to different Pods. _Other volume source will be supported soon._

The JSON schema of its value
```json
{
  "type": "object",
  "properties": {
    "volumes": {
      "type": "array",
      "items":{
        "description": "refer to https://kubernetes.io/docs/reference/kubernetes-api/config-and-storage-resources/volume/#Volume",
        "type": "object"
      }
    },
    "containers": {
      "type": "array",
      "items":{
        "description": "refer to https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#Container",
        "type": "object"
      }
    }
  }
}
```

## Installation

### Prerequisites
Spoditor depends on [Cert-Manager](https://cert-manager.io) to issue TLS certificates. So, please follow its [installation guide](https://cert-manager.io/docs/installation/kubernetes/) to install it to your Kubernetes cluster.

### Install Spoditor
Each Spoditor release offers an all-in-one YAML manifest `bundle.yaml`

```shell
kubectl apply -f https://github.com/spoditor/spoditor/releases/download/v0.1.1/bundle.yaml
```

## Quick Demo
[![asciicast](https://asciinema.org/a/xmA2TISTPQoMcXryyFnRiRxbI.svg)](https://asciinema.org/a/xmA2TISTPQoMcXryyFnRiRxbI)

## Contributing
We welcome pull request to support more ways Spoditor can argument the Pods of Statefulset.

Please refer to the [mount-volume](internal/annotation/volumes/mount.go) implementation to understand how to implement new annotation. Basically, all an annotation needs to do is to implement the following interfaces:
```go
type Handler interface {
	Mutate(spec *corev1.PodSpec, ordinal int, cfg interface{}) error
	GetParser() Parser
}

type Parser interface {
	Parse(annotations map[QualifiedName]string) (interface{}, error)
}
```