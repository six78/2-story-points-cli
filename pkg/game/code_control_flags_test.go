package game

func WithEnablePublishOnlineState(enable bool) Option {
	return func(game *Game) {
		game.codeControls.EnablePublishOnlineState = enable
	}
}
