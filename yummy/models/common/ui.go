package common

import (
	"github.com/GarroshIcecream/yummy/yummy/consts"
	tea "github.com/charmbracelet/bubbletea"
)

type TUIModel interface {
	tea.Model
	GetSessionState() consts.SessionState
	GetModelState() consts.ModelState
	SetSize(width, height int)
	GetSize() (width, height int)
}
