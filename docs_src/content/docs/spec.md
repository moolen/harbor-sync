# Spec

## SyncConfig

This is the root-level type.

```go
type HarborSync struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HarborSyncSpec   `json:"spec,omitempty"`
}
```

### HarborSyncSpec

ProjectSelector specifies how to find projects in harbor and how to map those to secrets in namespaces.
The `robotAccountSuffix` field defines what names the robot accounts have. The robot accounts always have a prefix of `robot$` - this is behavior is enforced by Harbor and might change in the future.

**Note:** The robot account suffix **should** be unique per `HarborSync`. If you map projects twice using two different `HarborSync` configurations you end up with a race condition.

```go
// HarborSyncSpec defines the desired state
// how should harbor projects map to secrets in namespaces
type HarborSyncSpec struct {

	// Specifies how to do matching on a harbor project.
	// Valid values are:
	// - "Regex" (default): interpret the project name as regular expression;
	Type ProjectMatchingType `json:"type"`

	// ProjectName specifies the project name
	ProjectName string `json:"name"`

	// PushAccess allows the robot account to push images, too. defaults to false.
	// As of now we can not tell what permissions a robot account has. The user
	// has to wait for the next rotation until the robot account has the new permissions.
	// Alternatively, you can re-create your HarborSync spec. This forces a rotation.
	PushAccess bool `json:"pushAccess"`

	// The RobotAccountSuffix specifies the suffix to use when creating a new robot account
	// +kubebuilder:validation:MinLength=4
	RobotAccountSuffix string `json:"robotAccountSuffix"`

	// The Mapping contains the mapping from project to a secret in a namespace
	Mapping []ProjectMapping `json:"mapping,omitempty"`

	// Webhook contains a list of endpoints which will be called
	// if the robot account changes (e..g automatic rotation, expired account, disabled...)
	// +optional
	Webhook []WebhookConfig `json:"webhook,omitempty"`
}
```

### ProjectMapping

ProjectMapping defines how to lookup namespaces in the cluster. Generally there are two lookup types: `Translate` and `Match`.

```go
// ProjectMapping defines how projects are mapped to secrets in specific namespaces
type ProjectMapping struct {
	Type      MappingType `json:"type"`
	Namespace string      `json:"namespace"`
	Secret    string      `json:"secret"`
}

// MappingType specifies how to map the project into the namespace/secret
// Only one of the following matching types may be specified.
// If none of the following types is specified, the default one
// is Translate.
// +kubebuilder:validation:Enum=Translate;Match
type MappingType string

const (
	// TranslateMappingType interpolates the project expression into the namespace
	TranslateMappingType MappingType = "Translate"

	// MatchMappingType treats the Namespace as regular expression and injects secrets into
	// all matching namespaces
	MatchMappingType MappingType = "Match"
)
```


### Webhook

Webhooks can be configured which will be called if the robot account credentials change. The only supported protocol is HTTP for now. Integrating other protocols is out of scope of this project. You should implement your own services that do the plumbing.

```go
// WebhookConfig defines how to call a webhook
type WebhookConfig struct {
	// Endpoint is a url
	Endpoint string `json:"endpoint"`
}

// WebhookUpdatePayload ...
type WebhookUpdatePayload struct {
	Project     string                 `json:"project"`
	Credentials RobotAccountCredential `json:"credentials"`
}

// RobotAccountCredential holds the robot account name & token to access the harbor API
type RobotAccountCredential struct {
	Name      string `json:"name"`
	CreatedAt int64  `json:"created_at"`
	Token     string `json:"token"`
}
```
