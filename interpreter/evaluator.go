package interpreter

type Evaluator struct {
	source      string
	filename    string
	projectRoot string
	modules     *moduleState

	runtime     *runtimeState
	currentTask *Task
}

func NewEvaluator() *Evaluator {
	return &Evaluator{modules: newModuleState(), runtime: newRuntimeState()}
}

func NewEvaluatorWithSource(source string) *Evaluator {
	return &Evaluator{source: source, modules: newModuleState(), runtime: newRuntimeState()}
}

func NewEvaluatorWithSourceAndFilename(source string, filename string) *Evaluator {
	return &Evaluator{source: source, filename: filename, modules: newModuleState(), runtime: newRuntimeState()}
}

func NewEvaluatorWithSourceFilenameAndRoot(source string, filename string, root string) *Evaluator {
	return &Evaluator{
		source:      source,
		filename:    filename,
		projectRoot: root,
		modules:     newModuleState(),
		runtime:     newRuntimeState(),
	}
}

func (e *Evaluator) SetProjectRoot(root string) {
	e.projectRoot = root
}

func (e *Evaluator) SetTaskFailurePolicy(policy string) error {
	if e.runtime == nil {
		e.runtime = newRuntimeState()
	}
	return e.runtime.setTaskFailurePolicy(policy)
}

func (e *Evaluator) cloneForTask(task *Task) *Evaluator {
	return &Evaluator{
		source:      e.source,
		filename:    e.filename,
		projectRoot: e.projectRoot,
		modules:     e.modules,
		runtime:     e.runtime,
		currentTask: task,
	}
}

func (e *Evaluator) newTask(parent *Task, internal bool) *Task {
	if e.runtime == nil {
		e.runtime = newRuntimeState()
	}
	t := newTask()
	t.internal = internal
	t.parent = parent
	t.source = e.source
	t.filename = e.filename
	if parent != nil {
		parent.addChild(t)
	}
	e.runtime.registerTask(t)
	return t
}
