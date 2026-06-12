package agent

import "fmt"

type Registry struct {
	agents map[string]Agent
}

func NewRegistry() *Registry {
	return &Registry{
		agents: make(map[string]Agent),
	}
}

func (r *Registry) Register(a Agent) {
	r.agents[a.Name()] = a
}

func (r *Registry) Get(name string) (Agent, error) {
	a, ok := r.agents[name]
	if !ok {
		return nil, fmt.Errorf("agent %q not registered", name)
	}
	return a, nil
}

func (r *Registry) Available() []Agent {
	var result []Agent
	for _, a := range r.agents {
		if ok, _ := a.Available(); ok {
			result = append(result, a)
		}
	}
	return result
}

func (r *Registry) All() []Agent {
	result := make([]Agent, 0, len(r.agents))
	for _, a := range r.agents {
		result = append(result, a)
	}
	return result
}
