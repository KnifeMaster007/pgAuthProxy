package utils

const (
	ConfigEnvPrefix             = "PGPROXY"
	ConfigAuthenticatorSection  = "authenticator"
	ConfigFileFlag              = "config"
	ConfigListenFlag            = "listen"
	ConfigCleartextPassword     = ConfigAuthenticatorSection + ".cleartext_password"
	ConfigCleartextPasswordFlag = "clear-passwd"
	ConfigCleartextPasswordEnv  = ConfigEnvPrefix + "_CLEAR_PASSWD"

	MetaPrefix                = "_META_"
	SourceMetaPrefix          = "_SOURCE_"
	SourceCredentialParameter = SourceMetaPrefix + "CRED"
	SourceSaltParameter       = SourceMetaPrefix + "SALT"
	TargetCredentialParameter = MetaPrefix + "TARGET_CRED"
	TargetHostParameter       = MetaPrefix + "TARGET_HOST"
)
