package app

import (
	"bufio"
	"fmt"
	"os/exec"
	"path/filepath"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/josegale/onering/internal/agent"
	"github.com/josegale/onering/internal/task"
	"github.com/josegale/onering/internal/terminal"
	"github.com/josegale/onering/internal/ui"
	"github.com/josegale/onering/internal/util"
)

type TaskOutput struct {
	mu  sync.Mutex
	buf string
}

func (o *TaskOutput) Append(s string) {
	o.mu.Lock()
	o.buf += s
	o.mu.Unlock()
}

func (o *TaskOutput) String() string {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.buf
}

type TaskRun struct {
	task.Task
	Sess   agent.Session
	Output *TaskOutput
	Status ui.TaskStatus
	IsPTY  bool
	cmd    *exec.Cmd
}

type TaskDoneMsg struct {
	Idx      int
	ExitCode int
}

func (m AppModel) taskItems() []ui.TaskItem {
	items := make([]ui.TaskItem, len(m.tasks))
	for i, t := range m.tasks {
		items[i] = ui.TaskItem{
			Name:      t.Name,
			Source:    string(t.Source),
			Dir:       t.Dir,
			Status:    t.Status,
			Preferred: m.state.IsPreferred(m.projectDir, t.Task.Key()),
		}
	}
	return items
}

func (m AppModel) activateTask(idx int, pty bool) (tea.Model, tea.Cmd) {
	if idx < 0 || idx >= len(m.tasks) {
		return m, nil
	}
	t := &m.tasks[idx]

	if t.Status == ui.TaskRunning {
		m.activeTask = idx
		m.activeApp = -1
		if t.IsPTY {
			layout, cmd := m.layout.SetActiveSession(t.Sess)
			m.layout = layout
			m.focus = ui.FocusMain
			m.layout = m.layout.SetFocus(ui.FocusMain)
			m = m.enterPassthrough()
			return m.syncSidebar(), cmd
		}
		tv := m.buildTaskView(idx)
		m.layout = m.layout.SetTaskView(&tv)
		m.focus = ui.FocusMain
		m.layout = m.layout.SetFocus(ui.FocusMain)
		return m.syncSidebar(), nil
	}

	// Clean up previous run if any
	if t.Sess != nil {
		m.layout = m.layout.RemoveSessionView(t.Sess.ID())
		t.Sess.Close()
		t.Sess = nil
	}
	t.Output = nil
	t.IsPTY = pty
	m.layout = m.layout.ClearTaskView()

	taskDir := m.projectDir
	if m.tasks[idx].Dir != "" {
		taskDir = filepath.Join(m.projectDir, m.tasks[idx].Dir)
	}

	if pty {
		sess, err := startTaskPTY(m.tasks[idx].Task, taskDir)
		if err != nil {
			return m, func() tea.Msg {
				return ErrorMsg{Err: fmt.Errorf("starting task %s: %w", t.Name, err)}
			}
		}
		t.Sess = sess
		t.Status = ui.TaskRunning
		m.activeTask = idx
		m.activeApp = -1
		layout, cmd := m.layout.SetActiveSession(sess)
		m.layout = layout
		m.focus = ui.FocusMain
		m.layout = m.layout.SetFocus(ui.FocusMain)
		m = m.enterPassthrough()
		return m.syncSidebar(), cmd
	}

	// Piped execution
	t.Output = &TaskOutput{}
	t.Status = ui.TaskRunning
	m.activeTask = idx
	m.activeApp = -1
	tv := m.buildTaskView(idx)
	m.layout = m.layout.SetTaskView(&tv)
	m.focus = ui.FocusMain
	m.layout = m.layout.SetFocus(ui.FocusMain)

	cmd := m.startTaskPiped(idx)
	return m.syncSidebar(), cmd
}

func (m AppModel) buildTaskView(idx int) ui.TaskViewModel {
	t := m.tasks[idx]
	label := string(t.Source) + ": " + t.Name
	tv := ui.NewTaskViewModel(label)
	w, h := m.layout.MainPanelTermSize()
	tv = tv.SetSize(w, h)
	if t.Output != nil {
		tv = tv.Append(t.Output.String())
	}
	return tv
}

func (m AppModel) startTaskPiped(idx int) tea.Cmd {
	t := &m.tasks[idx]
	cmd := exec.Command("sh", "-c", t.Command)
	cmd.Dir = m.projectDir
	if t.Dir != "" {
		cmd.Dir = filepath.Join(m.projectDir, t.Dir)
	}
	t.cmd = cmd

	stdout, _ := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout // merge stderr into stdout

	if err := cmd.Start(); err != nil {
		t.Status = ui.TaskFailed
		return nil
	}

	return func() tea.Msg {
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)
		for scanner.Scan() {
			line := scanner.Text() + "\n"
			t.Output.Append(line)
		}
		exitCode := 0
		if err := cmd.Wait(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			} else {
				exitCode = 1
			}
		}
		return TaskDoneMsg{Idx: idx, ExitCode: exitCode}
	}
}

func startTaskPTY(t task.Task, dir string) (agent.Session, error) {
	cmd := exec.Command("sh", "-c", t.Command)
	cmd.Dir = dir
	ptyHandle, err := terminal.StartPTY(cmd)
	if err != nil {
		return nil, err
	}
	label := string(t.Source) + ": " + t.Name
	return agent.NewPTYSession(ptyHandle, string(t.Source), label, util.NewID(), ""), nil
}

func (m AppModel) killTask(idx int) (tea.Model, tea.Cmd) {
	if idx < 0 || idx >= len(m.tasks) {
		return m, nil
	}
	t := &m.tasks[idx]
	if t.Sess != nil {
		id := t.Sess.ID()
		t.Sess.Close()
		t.Sess = nil
		m.layout = m.layout.RemoveSessionView(id)
	}
	if t.cmd != nil && t.cmd.Process != nil {
		t.cmd.Process.Kill()
		t.cmd = nil
	}
	t.Status = ui.TaskPending
	t.Output = nil

	if m.activeTask == idx {
		m.activeTask = -1
		m.layout = m.layout.ClearTaskView()
		if m.mode == ModePassthrough {
			m = m.exitToNavigation()
		}
		layout, cmd := m.layout.SetActiveSession(m.activeSession())
		m.layout = layout
		return m.syncSidebar(), cmd
	}
	return m.syncSidebar(), nil
}

func (m AppModel) refreshTasks(recursive bool) AppModel {
	var scanned []task.Task
	if recursive {
		scanned = task.ScanTasksRecursive(m.projectDir, m.config.Tasks.PackageManager)
	} else {
		scanned = task.ScanTasks(m.projectDir, m.config.Tasks.PackageManager)
	}
	m.state.SaveProjectTasks(m.projectDir, task.TasksToStored(scanned))

	preferred := m.state.PreferredTasks[m.projectDir]
	sorted := task.SortWithPreferred(scanned, preferred)

	running := make(map[string]*TaskRun)
	for i := range m.tasks {
		if m.tasks[i].Status == ui.TaskRunning {
			running[m.tasks[i].Task.Key()] = &m.tasks[i]
		}
	}

	newTasks := make([]TaskRun, len(sorted))
	for i, t := range sorted {
		if prev, ok := running[t.Key()]; ok {
			newTasks[i] = *prev
		} else {
			newTasks[i] = TaskRun{Task: t, Status: ui.TaskPending}
		}
	}
	m.tasks = newTasks
	return m
}

func (m AppModel) reorderTasks() AppModel {
	preferred := m.state.PreferredTasks[m.projectDir]
	// Extract the underlying tasks
	raw := make([]task.Task, len(m.tasks))
	for i := range m.tasks {
		raw[i] = m.tasks[i].Task
	}
	sorted := task.SortWithPreferred(raw, preferred)

	// Build a lookup from the old slice
	old := make(map[string]TaskRun)
	var activeKey string
	for i, t := range m.tasks {
		old[t.Task.Key()] = t
		if m.activeTask == i {
			activeKey = t.Task.Key()
		}
	}

	newTasks := make([]TaskRun, len(sorted))
	m.activeTask = -1
	for i, t := range sorted {
		if prev, ok := old[t.Key()]; ok {
			newTasks[i] = prev
		} else {
			newTasks[i] = TaskRun{Task: t, Status: ui.TaskPending}
		}
		if t.Key() == activeKey {
			m.activeTask = i
		}
	}
	m.tasks = newTasks
	return m
}

func (m AppModel) handleTaskTermError(msg terminal.TermErrorMsg) (AppModel, bool) {
	for i := range m.tasks {
		if m.tasks[i].Sess == nil || m.tasks[i].Sess.ID() != msg.ID {
			continue
		}
		m.tasks[i].Status = ui.TaskCompleted
		if m.activeTask == i && m.mode == ModePassthrough {
			m = m.exitToNavigation()
		}
		return m, true
	}
	return m, false
}
