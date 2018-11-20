# ocelot chart 

[ocelot](https://github.com/shankj3/ocelot/wiki) is a c.i./c.d. solution. this chart will deploy the admin, hookhandler, and poller components of the ocelot stack. 

 ## Chart Details 
 Installing this helm chart will culminate in 3 deployments:
  - hookhandler
  - admin
  - poller 
With service resources and (if enabled) ingress resources. 


## Configuration
 
### General 
| Parameter | Description | Default |
|:----------|-------------|--------:|
|`loglevel` | level of logging to set for all services | info |
| `serviceAccount` | service account to attach to deployments | default |
| `vault.protocol` | http or https | http |
| `vault.ip` | location of vault | `10.1.72.190` |
| `vault.port` | port that vault is listening to | 8200 |
| `secretName` | name of secret that has the vault token (if not using service account vault/kubernetes auth) | vault-token |
| `secretKey` | key to secret map that will return the vault token | token |
| `nsq.nsqlookupd.ip` | ip of nsqlookupd instance. will be connected to by all services for listening for nsq events | nsqlookupd-svc.default |
| `nsq.nsqlookupd.port` | port of nsqlookupd instance | 4161 |
| `nsq.nsqd.port` | port that nsqd is listening on. this chart assumes that nsqd is running as a daemonset with all kubernetes worker nodes exposing the nsqd service | 4150 |

### vault + kubernetes authentication
If this section is enabled, then each service will have an init container and sidecar attached to it. The init container will hold the [token-grinch](https://bitbucket.org/level11consulting/token-grinch/src/master/) code, which is a script that picks up the kubernetes service account token and uses it to authenticate with vault, then stores the received token at a filepath (`/etc/vaulted/token`). The sidecar container runs using the [token-renewer code](https://bitbucket.org/level11consulting/token-renewer/src), which will make sure that the token created in the init container does not expire. All the ocelot services will read the token from `/etc/vaulted/token`. 

configuration for utilizing the [kubernetes auth method](https://www.vaultproject.io/docs/auth/kubernetes.html)
  
| Parameter | Description | Default |
|:----------|-------------|--------:|
| `k8sToken.enabled` | whether or not to use vault + kubernetes authentication | false |
| `k8sToken.vaultRole` | role that is configured in vault for the current namespace that return a token valid for the ocelot vault policy | ocelot |
| `k8sToken.mountPath` | path to mount the authenticated token to in the other containers (**do not change**) | /etc/vaulted |
| `k8sToken.tokenFileName` | name of the actual file that holds just the authenticated token | token |
| `k8sToken.increment` | amount to renew the lease for | 10m |
| `k8sToken.sidecar.image` | image name for the sidecar to run alongside the ocelot sergvices to keep the token up to date | docker.metaverse.l11.com/token-renewer |
| `k8sToken.init.image` | image name for the initial init container that will take the service account and exchange it for an auth'd token | docker.metaverse.l11.com/token-grinch |


### admin-specific 

| Parameter | Description | Default |
|:----------|-------------|--------:|
| `admin.enabled` | whether or not to deploy admin service | true |
| `admin.swagger` | whether or not to serve swagger documentation | false |
| `admin.replicaCount` | number of replicas for admin service | 3 |
| `admin.image.repository` | docker repo of admin docker image | `docker.metaverse.l11.com/ocelot-admin` |
| `admin.image.tag` | docker tag of admin docker image | latest |
| `admin.image.pullPolicy` | pull policy for kubernetes for the admin image | Always | 
| `admin.service.type` | do not change --- should always be nodeport for now | NodePort |
| `admin.service.port` | where the admin container is listening for grpc connections. should not really change unless this is explictly changed on the admin container | 10000 |
| `admin.service.nodePort` | where on the kubernetes nodes to expose the admin service via NodePort | 31000 |
| `admin.ingress.enabled` | whether to create an ingress resource for GRPC-GATEWAY (ie for exposing rest API) | false |
| `admin.ingress.hosts[0]` | what name to use for ingress | ocelot-admin.metaverse.l11.com |
| `admin.ingress.tls[0].secretName` | tls secret for ingress | `metaverse-secret` | 
| `admin.ingress.tls[0].hosts[0]` | same as `admin.ingress.hosts[0]` | ocelot-admin.metaverse.l11.com |
| `admin.grpcIngress.enabled` | whether or not to create ingress resource for grpc (for the ocelot client) | false |
| `admin.grpcIngress.host` | name for grpc ingress resource | `ocelot-admin-grpc.metaverse.l11.com` |
| `admin.grpcIngress.tlsSecret` | tls secret for grpc ingress resource | metaverse-secret | 


### hookhandler-specific 

| Parameter | Description | Default |
|:----------|-------------|--------:|
| `hookhandler.enabled` | whether or not to deploy hookhandler service | true |
| `hookhandler.replicaCount` | number of replicas for hookhandler service | 3 |
| `hookhandler.image.repository` | docker repo for hookhandler image | docker.metaverse.l11.com/ocelot-hookhandler |
| `hookhandler.image.tag` | docker image tag for hookhandler | latest |
| `hookhandler.image.pullPolicy` | kubernetes pull policy for hookhandler service | Always |
| `hookhandler.service.type` | service type for hookhandler service. shouldn't change | ClusterIP |
| `hookhandler.service.port` | port that hookhandler is running on in the container, rarely will change | 8088 |
| `hookhandler.ingress.enabled` |  whether to create an ingress resource for hookhandler | false |
| `hookhandler.ingress.{x}` | follows normal kubernetes ingress spec | - |


### poller-specific
 
| Parameter | Description | Default |
|:----------|-------------|--------:|
| `poller.enabled` | whether or not to deploy poller service | true |
| `poller.image.repository` | docker repo for poller image | docker.metaverse.l11.com/ocelot-poller |
| `poller.image.tag` | docker image tag for poller image | latest |
| `poller.image.pullPolicy` | kubernetes pull policy | Always |
