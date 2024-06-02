package game

type gameConfig struct {
	EnableSymmetricEncryption bool
}

var defaultConfig = gameConfig{
	EnableSymmetricEncryption: true,
}
