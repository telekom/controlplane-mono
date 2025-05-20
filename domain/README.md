<!--
SPDX-FileCopyrightText: 2024 Deutsche Telekom AG

SPDX-License-Identifier: CC0-1.0    
-->

<p align="center">
  <h1 align="center">Domain</h1>
</p>

<p align="center">
  Some fancy sentence, maybe...
</p>

<p align="center">
  <a href="#reconciliation"> Reconciliation Flow</a> •
  <a href="#dependencies">Dependencies</a> •
  <a href="#model">Model</a> •
  <a href="#crds">CRDs</a>
</p>

## About

Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.
Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.  


## Reconciliation Flow
The diagram below shows the general Reconciliation flow.
# ![Flow](./docs/img/Organization_Operator.drawio.svg)


### Workflow
Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.
Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.
Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.
Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.

## Model
Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua.
These models are:
- [Subscription](https://github.com/telekom/pubsub-horizon-spring-parent/blob/main/horizon-core/src/main/java/de/telekom/eni/pandora/horizon/kubernetes/resource/Subscription.java): Represents a subscription of a subscriber to a specific event type. This subscription is used to filter and deliver event messages to the subscriber.
- [PublishedEventMessage](https://github.com/telekom/pubsub-horizon-spring-parent/blob/main/horizon-core/src/main/java/de/telekom/eni/pandora/horizon/model/event/PublishedEventMessage.java): Represents an event message that is published by a publisher.
- [SubscriptionEventMessage](https://github.com/telekom/pubsub-horizon-spring-parent/blob/main/horizon-core/src/main/java/de/telekom/eni/pandora/horizon/model/event/SubscriptionEventMessage.java): Represents an event message that is multiplexed from a [PublishedEventMessage](https://github.com/telekom/pubsub-horizon-spring-parent/blob/main/horizon-core/src/main/java/de/telekom/eni/pandora/horizon/model/event/PublishedEventMessage.java) by Galaxy for each subscriber.
- [Status](https://github.com/telekom/pubsub-horizon-spring-parent/blob/main/horizon-core/src/main/java/de/telekom/eni/pandora/horizon/model/event/Status.java): Represents the status of a [SubscriptionEventMessage](https://github.com/telekom/pubsub-horizon-spring-parent/blob/main/horizon-core/src/main/java/de/telekom/eni/pandora/horizon/model/event/SubscriptionEventMessage.java).
   <details>
     <summary>Status flow of a SubscriptionEventMessage</summary>

     ```mermaid
       graph TD;
         PROCESSED-->DELIVERING;
         PROCESSED-->FAILED;
         PROCESSED-->DROPPED;
         DELIVERING-->FAILED;
         DELIVERING-->DELIVERED;
         DELIVERING-->WAITING;
         PROCESSED-->WAITING;
     ```
  </details>
- [State](https://github.com/telekom/pubsub-horizon-spring-parent/blob/main/horizon-core/src/main/java/de/telekom/eni/pandora/horizon/model/db/State.java): Represents the state of an event message in the database. Contains timestamps, Kafka location information, filter results, errors, the status and additional metadata like tracing etc.
  <details>
    <summary>Example</summary>
  
    ```json
    {
      "_id": "410eacd1-0fc8-4718-b4cb-c8cf25baeb99",
      "event": {
        "id": "ede6cd87-14d2-4058-8186-f7937bbbdae7",
        "time": "2023-10-24T11:00:36.531Z",
        "type": "ecommerce.shop.orders.v1",
        "_id": "ede6cd87-14d2-4058-8186-f7937bbbdae7"
      },
      "coordinates": {
        "partition": 15,
        "offset": 50678896
      },
      "deliveryType": "CALLBACK",
      "environment": "playground",
      "eventRetentionTime": "DEFAULT",
      "modified": {
        "$date": {
          "$numberLong": "1707984041737"
        }
      },
      "multiplexedFrom": "d32f1150-2978-4641-8ebf-dfcd2b276071",
      "properties": {
        "X-B3-ParentSpanId": "77b58aa703c8e12a",
        "X-B3-Sampled": "1",
        "X-B3-SpanId": "c2a630bd02af829a",
        "X-B3-TraceId": "246db1ad668a55b269929ee9e1d1747f",
        "callback-url": "https://billing-service.example.com/api/v1/callback",
        "selectionFilterResult": "NO_FILTER",
        "subscriber-id": "ecommerce--billing--order-event-consumer"
      },
      "status": "WAITING",
      "subscriptionId": "4ca708e09edfb9745b1c9ceeb070aacde42cf04f",
      "timestamp": {
        "$date": {
          "$numberLong": "1707984041704"
        }
      },
      "topic": "subscribed"
    }
    ```
  </details>

## CRDs
Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.
You can find the custom resource definition here: [resources/crds.yaml](./resources/crds.yaml).

### Resource1
Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.

A simple example Resource1 would look like this:


<details>
  <summary>Example Resource1</summary>

  ```yaml
  apiVersion: a.b.c/v1
  kind: Resource1
  metadata:
    name: 4ca708e09edfb9745b1c9ceeb070aacde42cf04f
    namespace: prod
  spec:
  ```
</details><br />

### Resource2
Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.

A simple example Resource2 would look like this:


<details>
  <summary>Example Resource2</summary>

  ```yaml
  apiVersion: a.b.c/v1
  kind: Resource1
  metadata:
    name: 4ca708e09edfb9745b1c9ceeb070aacde42cf04f
    namespace: prod
  spec:
  ```
</details><br />

## Code of Conduct

This project has adopted the [Contributor Covenant](https://www.contributor-covenant.org/) in version 2.1 as our code of conduct. Please see the details in our [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md). All contributors must abide by the code of conduct.

By participating in this project, you agree to abide by its [Code of Conduct](./CODE_OF_CONDUCT.md) at all times.

## Licensing

This project follows the [REUSE standard for software licensing](https://reuse.software/). You can find a guide for developers at https://telekom.github.io/reuse-template/.   
Each file contains copyright and license information, and license texts can be found in the [./LICENSES](./LICENSES) folder. For more information visit https://reuse.software/.