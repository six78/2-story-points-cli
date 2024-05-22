package view

import (
	"2sp/internal/config"
	"2sp/internal/transport"
	"2sp/pkg/game"
	tea "github.com/charmbracelet/bubbletea"
	"go.uber.org/zap"
)

func Run(game *game.Game, transport transport.Service) int {
	m := initialModel(game, transport)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		config.Logger.Error("error running program", zap.Error(err))
		return 1
	}
	return 0
}
