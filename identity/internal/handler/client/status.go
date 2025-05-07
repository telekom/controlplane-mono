package client

import (
	"github.com/telekom/controlplane-mono/common/pkg/condition"

	identityv1 "github.com/telekom/controlplane-mono/identity/api/v1"
)

var (
	// Processing
	processingCondition = condition.NewProcessingCondition("ClientProcessing",
		"Processing client")
	processingNotReadyCondition = condition.NewNotReadyCondition("ClientNotReady",
		"Client not ready")

	// Blocked
	blockedCondition         = condition.NewBlockedCondition("Realm not found")
	blockedNotReadyCondition = condition.NewNotReadyCondition("RealmNotFound", "Realm not found")

	// Waiting
	waitingCondition = condition.NewProcessingCondition("ClientProcessing",
		"Waiting for Realm to be processed")
	waitingNotReadyCondition = condition.NewNotReadyCondition("ClientProcessing",
		"Waiting for Realm to be processed")

	// Ready
	doneProcessingCondition = condition.NewDoneProcessingCondition("Created Client")
	readyCondition          = condition.NewReadyCondition("Ready", "Client is ready")
)

func MapToClientStatus(realmStatus *identityv1.RealmStatus) identityv1.ClientStatus {
	return identityv1.ClientStatus{
		IssuerUrl: realmStatus.IssuerUrl,
	}
}

func SetStatusProcessing(currentStatus *identityv1.ClientStatus, client *identityv1.Client) {
	client.Status = *currentStatus
	client.SetCondition(processingCondition)
	client.SetCondition(processingNotReadyCondition)
}

func SetStatusBlocked(currentStatus *identityv1.ClientStatus, client *identityv1.Client) {
	client.Status = *currentStatus
	client.SetCondition(blockedCondition)
	client.SetCondition(blockedNotReadyCondition)
}

func SetStatusWaiting(currentStatus *identityv1.ClientStatus, client *identityv1.Client) {
	client.Status = *currentStatus
	client.SetCondition(waitingCondition)
	client.SetCondition(waitingNotReadyCondition)
}

func SetStatusReady(currentStatus *identityv1.ClientStatus, client *identityv1.Client) {
	client.Status = *currentStatus
	client.SetCondition(doneProcessingCondition)
	client.SetCondition(readyCondition)
}
