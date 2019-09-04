# Spec

## SyncConfig

This is the root-level type.

```go
type HarborSync struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HarborSyncSpec   `json:"spec,omitempty"`

	// Status stores the plain credentials of the robot accounts
	// this is necessary, because the credentials are not accessible via API
	Status HarborSyncStatus `json:"status,omitempty"`
}
```

### HarborSyncSpec

ProjectSelector specifies how to find projects in harbor and how to map those to secrets in namespaces.
The `robotAccountSuffix` field defines what names the robot accounts have. The robot accounts always have a prefix of `robot$` - this is behavior is enforced by Harbor and might change in the future.

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

	// TODO:
	// AccessType: ReadOnly|ReadWrite

	// The RobotAccountSuffix specifies the suffix to use when creating a new robot account
	// +kubebuilder:validation:MinLength=4
	RobotAccountSuffix string `json:"robotAccountSuffix"`

	// The Mapping contains the mapping from project to a secret in a namespace
	Mapping []ProjectMapping `json:"mapping"`

	// Webhook is WIP
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
