package commands

// NewDefaultRegistry creates a registry with all default commands
func NewDefaultRegistry() *CommandRegistry {
	registry := NewCommandRegistry()
	
	// register all default commands
	registry.Register(&ClearCommand{})
	registry.Register(&ContextCommand{})
	registry.Register(&FocusCommand{})

	registry.Register(&WorkspaceCommand{})
	registry.Register(&BookmarkCommand{})
	registry.Register(&ModelCommand{})

	registry.Register(&VerbosityCommand{})
	registry.Register(&HelpCommand{})
	
	return registry
}
