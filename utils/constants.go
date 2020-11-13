package utils

const (
	FlagConfigFile = "config"
	FlagListen     = "listen"

	MetaPrefix                = "_META_"
	SourceMetaPrefix          = "_SOURCE_"
	SourceCredentialParameter = SourceMetaPrefix + "CRED"
	SourceSaltParameter       = SourceMetaPrefix + "SALT"
	TargetCredentialParameter = MetaPrefix + "TARGET_CRED"
	TargetHostParameter       = MetaPrefix + "TARGET_HOST"
)
