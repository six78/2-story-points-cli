package game

type FeatureFlags struct {
	EnableDeckSelection bool
}

func defaultFeatureFlags() FeatureFlags {
	return FeatureFlags{
		EnableDeckSelection: false,
	}
}
