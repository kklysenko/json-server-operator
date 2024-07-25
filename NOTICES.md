> After I went over the task definition, I've decided to sketch a design to have a roadmap and avoid
loosing details.

# Design

## Introduction

Aim of the solution is to automate the deployment and management of JSON-server instances in 
a Kubernetes environment. The main goal is to provide users with a simple and efficient way to 
configure, deploy, and scale JSON-server instances.

## Functional requirements

- Users should be able to create an JsonServer object in the Kubernetes API which will represent
  a json-server instance they want to create
  - As a result of JsonServer creating users should:
      - have a new Deployment
      - have a Service to expose the Deployment
      - have a ConfigMap to store the json config which will be mounted into the Deployment
- Users should be able to specify the replicas to use
- Users should be able to specify json config their application needs.
- Users shouldn't be able to create JsonServer with a name that has no "app-" prefix
- Users shouldn't be able to create JsonServer with a valid JSON object
- Users should be able to modify JsonServer resource
  - Result of JsonServer modification should be reflected in respective Deployment, Service and ConfigMap
- Users should be able to delete all respective entities attached to JsonServer by deleting JsonServer
- Users should be able to view JsonServer status indicating the object was synced successfully or no
- Users should be able to access the json-server instance using respective Service and see the json config they specified 
in the JsonServer object

## Non-functional requirements

1. Reliability: System should handle failures gracefully, providing clear error messages.
2. Performance: System should be optimized for efficient resource usage and quick response times.
3. Flexibility: System should provide an open-ended way to configure resources.
4. Reusability: System should be easily shared between Kubernetes clusters.
5. Usability: System should offer intuitive user interactions.
6. Scalability: System should support horizontal scaling to manage increased workloads efficiently.
7. Resilience: System should maintain stability under varying loads and network conditions.
8. Observability: System should implement logging to troubleshoot issues.
9. Compliance: System should follow the Kubernetes standards for operator design and operation.

## Components

1. Custom Resource Definition (CRD)
2. Controller
3. Admission Webhooks

### Custom Resource Definition (CRD)

#### Field Types and Descriptions
Assumed we don't eed to describe common resource fields as apiVersion, kind, spec etc.

The JsonServer resource required having next common fields:
- apiVersion
- kind
- metadata
- spec
- status

> Here I'd add the spec.selector.matchLabels and spec.template.metadata fields to be able to set labels for Pods and
> give more flexibility. I decided that I will generate labels automatically based on the metadata.name field to avoid
> overcomplicating the solution

The JsonServer resource required having next specific fields:

- spec.replicas:
  Type: integer
  Description: The number of JSON-server instances to run.
  Constraints: Must be a positive integer (minimum: 1).
- spec.jsonConfig:
  Type: string
  Description: JSON configuration to be used by the JSON-server instance.
  Constraints: Must be a valid JSON object.

- status.state:
  Type: string
  Description: Indicates the synchronization state of the JsonServer resource.
  Possible Values:
    Synced
    Error
- status.message:
  Type: string
  Description: Provides additional details about the state, such as error messages or success notes.

#### JsonServer resource example

```yaml
apiVersion: example.com/v1
kind: JsonServer
metadata:
  name: app-my-server
  namespace: default
spec:
  replicas: 2
  jsonConfig: |
    {
      "people": [
        {
          "id": 1,
          "name": "Person A"
        },
        {
          "id": 2,
          "name": "Person B"
        }
      ]
    }
status:
  state: Synced
  message: "Synced succesfully!"
```

### Controller
The controller is responsible for managing the lifecycle of JsonServer resources in the Kubernetes cluster. 
It automates the creation, update, and deletion of associated Kubernetes resources based on the state of JsonServer 
objects.

#### Permissions
To interact with Kubernetes resources JsonServer requires to have the following permissions:
- apiGroups: [""]
  resources: ["services", "configmaps"]
  verbs: ["get", "list", "watch", "create", "update", "delete"]
- apiGroups: ["apps"]
  resources: ["deployments"]
  verbs: ["get", "list", "watch", "create", "update", "delete"]
- apiGroups: ["example.com"]
  resources: ["jsonservers"]
  verbs: ["get", "list", "watch", "create", "update", "delete"]
- apiGroups: ["example.com"]
  resources: ["jsonservers/status", jsonservers/finalizers]
  verbs: ["get", "update", "patch"]

#### Events to Handle

Assumed the JsonServer status is being updated on every event occurrence.

*JsonServer Created*
Actions:
  - Create a new Deployment
    - Set the number of replicas as specified in the JsonServer
    - Set the Deployment name and namespace
    - Set label selectors based on JsonServer name
    - Configure a container
    - Configure and mount a volume
  - Create a new Service
    - Set the Service to use port 3000 and target the Deployment's port 3000
  - Create a new ConfigMap
    - Store the JSON configuration with the key "db.json"
  - Set owner references on the Deployment, Service, and ConfigMap to ensure proper garbage collection when the JsonServer is deleted
  - Add finalizers to the JsonServer to ensure resources are not prematurely deleted and that cleanup tasks are performed before deletion

> I'd validate if child resources with same name and namespace are already exist and reflect it in the JsonServer status

*JsonServer Updated*
Actions:
  - Update the existing Deployment
    - Update the number of replicas if changed
  - Update the existing ConfigMap
    - Update the ConfigMap with the new JSON configuration from the JsonServer
    > I'd implement rollout triggering on CM change, but skipped as there was no requirement for that

*JsonServer Deleted*
Actions:
  - Delete the associated Deployment, Service, and ConfigMap

*Owned Resource Deleted*
Actions:
  - Detect and recreate the missing resource
  - Set owner reference on the recreated resource

*Owned Resource Modified*
  - Update the modified resource in accordance to JsonServer

### Admission Webhooks

#### Validation Webhook

The main idea of validation webhook is to ensure that any JsonServer custom resources created or updated adhere
to specific requirements before being accepted by the Kubernetes API server. The next validations should be applied to
JsonServer:

> Regarding the `spec.jsonConfig` field validation. I'd send the JSON configuration to the validation webhook to prevent setting an
> incorrect config for the resource, ensuring issues are caught early before they reach the reconciler. So in this case
> having "Error: spec.jsonConfig is not a valid json object" error in JsonServer status is impossible.

- `metadata.name`
  - Ensure the name follows the naming convention `app-${name}`
- `spec.jsonConfig`
  - Ensure it contains valid JSON
- `spec.replicas`
  - Ensure that the field is a non-negative integer

#### Mutating Webhook

Automatically sets default values for fields in JsonServer custom resources if not provided by the user.

> This was out of requirements but because I wanted JsonServer to be aligned with Deployment I decided that `spec.replicas`
> field should have a default value

- `spec.replicas`
  - If the field is not specified, set it to 1

## Development Plan

### Setup Environment

- Use Kind for local testing
- Use kubebuilder for controller creation
  > I've decided to not go with Operator SDk as it has default integration with OLM and overcomplicates the flow

### CRD and Controller Implementation
- Define JsonServer CRD schema
- Implement controller logic for resource management

### Admission Webhook
- Develop Validation webhook
- Develop Mutation webhook

### Testing

[//]: # (- Unit tests for controller logic)

[//]: # (- Integration tests using local Kubernetes cluster)
- Manual verification with kubectl


## Test Plan

TC-01 JsonServer Created
Action: Create a JsonServer with valid inputs.
Expected Outcome:
- Deployment, Service, and ConfigMap are created with correct names and namespaces.
- Deployment has correct labels and selectors.
- Deployment container uses the specified image and mounts volume at /data/db.json.
- All pods are created, have correct labels and running without errors.
- Service exposes Deployment on port 3000 with the correct selector and port configuration.
- ConfigMap contains JSON config under key "db.json", the content match expected structure
- JSON config is accessible via port-forwarding and matches expected data.
- The JsonServer `status.message` field is set to "Synced successfully!" and the `status.state` set to "Synced".

TC-02 Default Replica Value
Action: Create a JsonServer without replicas.
Expected Outcome:
- Deployment is created with 1 replica.

TC-03 Invalid Name
Action: Create a JsonServer without "app-" prefix.
Expected Outcome:
- Request is rejected with error message: "Invalid name: must start with 'app-'."

TC-04 Invalid JSON Config
Action: Create with malformed JSON config.
Expected Outcome:
- Request is rejected with error message: "Invalid JSON configuration."

TC-05 Update Replicas
Action: Modify replicas in an existing JsonServer.
Expected Outcome:
- Deployment updates to new replica count. Verify the updated replica count in Deployment.

TC-06 Update JSON Config
Action: Change JSON config in an existing JsonServer.
Expected Outcome:
- ConfigMap updates with new JSON. Validate content and check for correct sync status.

TC-07 Delete JsonServer
Action: Delete an existing JsonServer.
Expected Outcome:
- Associated Deployment, Service, and ConfigMap are deleted.

TC-08 Owned Resource Deleted
Action: Manually delete Deployment/Service/ConfigMap.
Expected Outcome:
- Operator recreates missing resources. 
- Resources match JsonServer specification.

TC-09 Owned Resource Modified
Action: Manually modify Deployment/Service/ConfigMap.
Expected Outcome:
- Operator reconciles to match JsonServer.
- Resources match JsonServer specification.

TC-10 Service Exposure
Action: Port-forward to the Service.
Expected Outcome:
- Service is accessible on port 3000. 
- Application is accessible via curl command or browser.

TC-11 Invalid Replica NUmber
Action: Create a JsonServer with negative number of replicas.
Expected Outcome:
- Request is rejected with error message: "Invalid replicas number."
