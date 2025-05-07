package v1

type FeatureType string

// Independent Features
const (
	FeatureTypePassThrough   FeatureType = "PassThrough"
	FeatureTypeAccessControl FeatureType = "AccessControl"
	FeatureTypeRateLimit     FeatureType = "RateLimit"
)

// Dependent Features
const (
	FeatureTypeLastMileSecurity FeatureType = "LastMileSecurity" // depends on AccessControl
	FeatureTypeExternalIDP      FeatureType = "ExternalIDP"      // depends on LastMileSecurity
	FeatureTypeCustomScopes     FeatureType = "CustomScopes"     // depends on LastMileSecurity
)
