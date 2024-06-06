package game

type FeatureFlags struct {
	EnableDeckSelection bool
}

func defaultFeatureFlags() FeatureFlags {
	return FeatureFlags{
		EnableDeckSelection: false,
	}
}

type codeControlFlags struct {
	EnablePublishOnlineState bool
}

func defaultCodeControlFlags() codeControlFlags {
	return codeControlFlags{
		EnablePublishOnlineState: true,
	}
}
