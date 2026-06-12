package ui

type PromptModel struct {
	width  int
	height int
}

func NewPromptModel() PromptModel {
	return PromptModel{}
}

func (m PromptModel) SetSize(w, h int) PromptModel {
	m.width = w
	m.height = h
	return m
}

func (m PromptModel) View() string {
	return MutedStyle.Render("› Press 'i' to enter a prompt")
}
