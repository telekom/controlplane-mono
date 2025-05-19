<!--
SPDX-FileCopyrightText: 2024 Deutsche Telekom AG

SPDX-License-Identifier: CC0-1.0    
-->

<p align="center">
  <img src="./docs/img/Open-Telekom-Integration-Platform_Visual.svg" alt="Open Telekom Integration Platform logo" width="200">
  <h1 align="center">Control Plane</h1>
</p>

<p align="center">
 A centralized management layer that maintains the desired state of your systems by orchestrating workloads, scheduling, and system operations through a set of core and custom controllers.
</p>

<p align="center">
  <a href="#features">Features</a> •
  <a href="#components"> Components</a> •
  <a href="#getting-started">Getting started</a>
</p>

## About

As Part of Open Telecom Integration Platform, the Control Plane is the central management layer that governs the operation of your Kubernetes cluster. It maintains the desired state of the system, manages workloads, and provides interfaces for user interaction and automation.

The Rover Control Plane components run on one or more nodes in the cluster and coordinate all cluster activities, including scheduling, monitoring, and responding to events.  


## Features

The Open Telekom Integration Platform Control Plane supports the whole API lifecycle and allows seamless, cloud-independent integration of services. Further, it enables a fine-grained and vigilant API access control. The communication is secure by design, utilizing OAuth 2.0 and an integrated permission management. 

Key features of The Rover Control Plane include:  


<details>
<summary><strong>API Management</strong></summary>  
Control Plane supports the whole API lifecycle and allows seamless, cloud-independent integration of services. Further, it enables a fine-grained and vigilant API access control. The communication is secure by design, utilizing OAuth 2.0 and an integrated permission management. 
</details>
<br />
<details>
<summary><strong>Approval Management</strong></summary>  
It provides secure and auditable access to APIs, with features like 4-eyes-principle, approval expiration, recertification, and more.
</details>
<br />
<details>
<summary><strong>Organization and Admin Mechanism</strong></summary>  
Provides Administrative tools for efficient organization management, including zones, gateways, and identity providers.
</details>
<br />
<details>
<summary><strong>Team Management:</strong></summary>  
Provides team management capabilities within the control plane
</details>
<br />
<details>
<summary><strong>Secret-Management</strong></summary>
Secret management involves securely storing, accessing, and distributing sensitive information such as passwords, API keys, and certificates within a Kubernetes cluster. It ensures that secrets are encrypted at rest and transmitted securely, while limiting access to only authorized workloads and users.  
</details>

<details>
<summary><strong>REST APIs for Key Actions</strong></summary> 
- Rover API: API to interact with and manage Rover functionalities.
- Approval API: API for handling approval processes and workflows.
- Team API: API for team management and related actions.
- Catalog API: API to access and manage an API catalog
</details>

## Components

### Controllers
In addition to the core components, the control plane may also run custom operators. These are specialized control loops designed to manage complex domain-specific applications and configurations. These operators extend Kubernetes functionality using the [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/), combining custom resource definitions (CRDs) with controllers that automate lifecycle management.

The following operators run on the control plane:
- Rover Operator
- Application Operator
- Admin Operator
- Organization Operator
- API Operator
- Gateway Operator
- Identity Operator
- Approval Operator

These operators work alongside the Kubernetes API server and etcd, watching for changes to custom resources and ensuring the actual state of their managed components aligns with the desired configuration.

### API Servers
- Secret Manager: RESTful API for managing secrets. It allows you to store, retrieve, and delete secrets securely.
- Rover-Server: RESTful API for managing Rover resources such as Applications, ApiSpecifications, ApiExposures, ApiSubscriptions
- Organization-Server
- Controlplane API

### Libraries
- Common
- Common-Server

### Infrastructure

Rover Control Plane requires the following infrastructure components in order to operate correctly:

- **Kubernetes**: The Rover Control Plane is designed to be deployed on Kubernetes. Currently, it is tested with Kubernetes version 1.31.


### Utilities



## Architecture
The diagram below shows the general flow and interfaces between the most important components of The Rover Control Plane.
# ![Architecture](.docs//img/CP_Architecture_2.drawio.svg)

### Workflow

## CRDs


### Rover resource
### Subscription resource
### ...

All subscription information is currently stored in "Subscription" custom resources which will be watched by the Horizon components.
You can find the custom resource definition here: [resources/crds.yaml](./resources/crds.yaml).

A simple example Subscription would look like this:


<details>
  <summary>Example Subscription</summary>

  ```yaml
  apiVersion: subscriber.horizon.telekom.de/v1
  kind: Subscription
  metadata:
    name: 4ca708e09edfb9745b1c9ceeb070aacde42cf04f
    namespace: prod
  spec:
    subscription:
      callback: >-
        https://billing-service.example.com/api/v1/callback
      deliveryType: callback
      payloadType: data
      publisherId: ecommerce--shop--order-events-provider
      subscriberId: ecommerce--billing--order-event-consumer
      subscriptionId: 4ca708e09edfb9745b1c9ceeb070aacde42cf04f
      trigger: {}
      type: ecommerce.shop.orders.v1
  ```
</details><br />


## Getting started

If you want to learn more about how to install and run the Rover Control Plane in a Kubernetes environment, visit: [Installing Rover Control Plane](./files/installation.md)  
But if you want to get started right away with a non-productive local environment and try out Rover Control Plane, we recommend visting: [Local installation (Quickstart)](./files/quickstart.md). 


## Code of Conduct

This project has adopted the [Contributor Covenant](https://www.contributor-covenant.org/) in version 2.1 as our code of conduct. Please see the details in our [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md). All contributors must abide by the code of conduct.

By participating in this project, you agree to abide by its [Code of Conduct](./CODE_OF_CONDUCT.md) at all times.

## Licensing

This project follows the [REUSE standard for software licensing](https://reuse.software/). You can find a guide for developers at https://telekom.github.io/reuse-template/.   
Each file contains copyright and license information, and license texts can be found in the [./LICENSES](./LICENSES) folder. For more information visit https://reuse.software/.
