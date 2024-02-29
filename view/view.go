package view

import (
	tea "github.com/charmbracelet/bubbletea"
	"go.uber.org/zap"
	"os"
	"waku-poker-planning/config"
	"waku-poker-planning/game"
)

type View struct {
	logger  *zap.Logger
	backend *game.Game
	model   model
}

func NewView(backend *game.Game) *View {
	return &View{
		logger:  config.Logger.Named("view"),
		backend: backend,
	}
}

func (v *View) Run() {
	m := initialModel(v.backend.CurrentState())
	m.stateChannel = v.backend.SubscribeToStateChanges()
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		v.logger.Error("error running program", zap.Error(err))
		os.Exit(1)
	}
	v.logger.Info("program finished")
}
