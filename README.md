# Helm-restful
A http server using `go-restful and Helm Go SDK that can execute helm command to target Kubernetes cluster with HTTP request api.

## Usage
```bash
Usage of ./helm-restful:
  -h, --host string                host to which the server will listen (default "0.0.0.0")
      --kubeconfig string          path to kubeconfig file of target cluster
  -p, --port string                port to which the server will listen (default "8080")
      --registry-config string     path to the registry config file
      --repository-cache string    path to the file containing cached repository indexes 
      --repository-config string   path to the file containing repository names and URLs
```
For example, after successfully build compile the package, user can run the server with command shown below:

```bash
$ ./helm-restful --kubeconfig </path/to/kubeconfig>
```
## APIs
| Path | Method | Discription |cc.MiddlerWare(middlewares.Logger)
| :----:| :----: | :----: |
| /helms | GET | List all helm releases from all namespaces |
| /helms/{namespace} | GET | List all helm releases from a target namespace |
| /helms/{namespace}/{release-name} | GET | Get info of target helm release from a target namespace |
| /helms/ | POST | Install helm chart |
| /helms/{namespace}/{release-name} | PUT | Upgrate target helm chart |
| /helms/{namespace}/{release-name} | DELETE | Delete target release from target namespace |

------


### List all releases
```
GET /helms
```

#### **Example**
Request
```bash
curl --location --request GET 'localhost:8080/helms/'
```

Response
```JSON
{
  "total_items": 3,
  "releases": [
    {
      "name": "mongodb",
      "status": "deployed",
      "namespace": "test2",
      "version": 1,
      "deployed_at": "2022-02-14 18:55:49",
      "description": "Install complete"
    },
    {
      "name": "mongodb",
      "status": "deployed",
      "namespace": "test",
      "version": 1,
      "deployed_at": "2022-02-14 18:55:49",
      "description": "Install complete"
    },
    {
      "name": "operator",
      "status": "deployed",
      "namespace": "test2",
      "version": 1,
      "deployed_at": "2022-02-14 16:11:54",
      "description": "Install complete"
    }
  ]
}
```
------


### List all releases of target namespace
#### **Request**

```
GET /helms/{namespace}
```
**Query Parameter**
| Name | Type| Required| Discription |
| :----:| :----: | :----: | :----: |
|namespace|string|true|Name of target namespace|
#### **Example**
Request
```bash
curl --location --request GET 'localhost:8080/helms/test'
```

Response
```JSON
{
  "total_items": 2,
  "releases": [
    {
      "name": "mongodb",
      "status": "deployed",
      "namespace": "test",
      "version": 1,
      "deployed_at": "2022-02-14 18:55:49",
      "description": "Install complete"
    },
    {
      "name": "operator",
      "status": "deployed",
      "namespace": "test2",
      "version": 1,
      "deployed_at": "2022-02-14 16:11:54",
      "description": "Install complete"
    }
  ]
}
```
------


### Retrieve target helm chart
```
GET /helms/{namespace}/{release-name}
``` 
**Query Parameter**

| Name | Type| Required| Discription |
| :----:| :----: | :----: | :----: |
|namespace|string|true|Name of target namespace|
|release-name|string|true|Name of release|


#### **Example**
Request
```bash
curl --location --request GET 'localhost:8080/helms/test/mongodb'
```

Response
```JSON
{
  "name": "mongodb",
  "status": "deployed",
  "namespace": "test",
  "version": 1,
  "deployed_at": "2022-02-14 18:55:49",
  "description": "Install complete"
}
```
------


### Install helm chart
#### **Request**

```
POST /helms/
```

**Request Body**
| Name | Type| Required| Default|Discription |
| :----:| :----: | :----: | :----: | :----: |
|name|string|true||Name of release|
|chart|string|true||Absolute path to target helm chart folder/archive or URL|
|namespace|string|false|default|Namespace scope for this request|
|values|string|false||Specify values in a YAML file or a URL|

#### **Example**
Request
```bash
curl --location --request POST 'localhost:8080/helms' \
--data-raw '{
    "name": "mongodb",
    "chart": "bitnami/mongodb",
    "namespace": "test",
    "values": {
        "architecture": "replicaset",
        "auth": {
            "rootPassword": "123456",
            "replicaSetKey": "test"
        }
    }
}'
```

Response
```JSON
{
  "name": "mongodb",
  "status": "deployed",
  "namespace": "test",
  "version": 1,
  "deployed_at": "2022-02-14 18:55:49",
  "description": "Install complete"
}
```
------


### Update helm chart
#### **Request**

```
PUT /helms/{namespace}/{release-name}
```
**Query Parameter**

| Name | Type| Required| Discription |
| :----:| :----: | :----: | :----: |
|namespace|string|true|Name of target namespace|
|release-name|string|true|Name of release|

**Request Body**
| Name | Type| Required| Default|Discription |
| :----:| :----: | :----: | :----: | :----: |
|chart|string|true||Absolute path to target helm chart folder/archive or URL|
|values|string|false||Specify values in a YAML file or a URL|

#### **Example**
Request
```bash
curl --location --request PUT 'localhost:8080/helms/test/mongodb' \
--data-raw '{
    "chart": "bitnami/mongodb",
    "values": {
        "architecture": "replicaset",
        "auth": {
            "rootPassword": "123456",
            "replicaSetKey": "test"
        },
        "replicaCount": 1
    }
}'
```

Response
```JSON
{
  "name": "mongodb",
  "status": "deployed",
  "namespace": "test",
  "version": 2,
  "deployed_at": "2022-02-14 18:55:49",
  "description": "Upgrade complete"
}
```

------

### Delete helm chart
#### **Request**

```
GET /helms/{namespace}/{release-name}
```

**Query Parameter**

| Name | Type| Required| Discription |
| :----:| :----: | :----: | :----: |
|namespace|string|true|Name of target namespace|
|release-name|string|true|Name of release|

#### **Example**
Request
```bash
curl --location --request DELETE 'localhost:8080/helms/test/mongodb'
```