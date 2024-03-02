package view

import (
	tea "github.com/charmbracelet/bubbletea"
	"go.uber.org/zap"
	"waku-poker-planning/app"
	"waku-poker-planning/config"
)

func Run(a *app.App) int {
	m := initialModel(a)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		config.Logger.Error("error running program", zap.Error(err))
		return 1
	}
	return 0
}
