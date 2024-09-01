package game

type FeatureFlags struct {
}

func defaultFeatureFlags() FeatureFlags {
	return FeatureFlags{}
}

type codeControlFlags struct {
	EnablePublishOnlineState bool
}

func defaultCodeControlFlags() codeControlFlags {
	return codeControlFlags{
		EnablePublishOnlineState: true,
	}
}
