package view

import (
	"2sp/internal/app"
	"2sp/internal/config"
	tea "github.com/charmbracelet/bubbletea"
	"go.uber.org/zap"
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
