package ui

type AgentViewModel struct {
	content string
	width   int
	height  int
}

func NewAgentViewModel() AgentViewModel {
	return AgentViewModel{}
}

func (m AgentViewModel) SetSize(w, h int) AgentViewModel {
	m.width = w
	m.height = h
	return m
}

func (m AgentViewModel) View() string {
	if m.content == "" {
		return MutedStyle.Render("Start a session and send a prompt to begin.")
	}
	return m.content
}
