package config

var (
	EnvironmentLabelKey = BuildLabelKey("environment")
	OwnerUidLabelKey    = BuildLabelKey("owner.uid")
)

func BuildLabelKey(key string) string {
	return LabelKeyPrefix + "/" + key
}
