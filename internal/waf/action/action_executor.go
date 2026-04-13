package action

import "log"

type Executor struct {
	registry *Registry
}

func NewExecutor(registry *Registry) *Executor {
	return &Executor{registry: registry}
}

func (e *Executor) ExecuteAll(actionNames []string, ap *Params) bool {
	shouldBlock := false
	for _, name := range actionNames {
		if name == BlockRequestActionName {
			shouldBlock = true
			continue
		}
		action, ok := e.registry.Get(name)
		if !ok {
			// TO DO - логировать в файл
			continue
		}
		action.Do(ap)
	}
	return shouldBlock
}

func (e *Executor) Execute(actionName string, ap *Params) bool {
	if actionName == BlockRequestActionName {
		return true
	}
	action, ok := e.registry.Get(actionName)
	if !ok {
		log.Printf("failed to find action %s in base\n", actionName)
		return false
	}
	action.Do(ap)
	return false
}
