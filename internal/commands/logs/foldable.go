package logs

import (
	"github.com/cirruslabs/echelon"
	"github.com/cirruslabs/echelon/renderers"
)

// Foldable log renderer prints start and end messages when a scope is started and finished respectively.
type FoldableLogsRenderer struct {
	delegate          *renderers.SimpleRenderer
	startFoldTemplate string
	endFoldTemplate   string
	escapeFunc        func(s string) string

	renderers.StubRenderer
}

func (r FoldableLogsRenderer) RenderScopeStarted(entry *echelon.LogScopeStarted) {
	if !r.delegate.ScopeHasStarted(entry.GetScopes()) {
		r.printFoldMessage(entry.GetScopes(), r.startFoldTemplate)
	}
	r.delegate.RenderScopeStarted(entry)
}

func (r FoldableLogsRenderer) RenderScopeFinished(entry *echelon.LogScopeFinished) {
	r.delegate.RenderScopeFinished(entry)
	r.printFoldMessage(entry.GetScopes(), r.endFoldTemplate)
}

func (r FoldableLogsRenderer) RenderMessage(entry *echelon.LogEntryMessage) {
	r.delegate.RenderMessage(entry)
}

func (r FoldableLogsRenderer) RenderRawMessage(message string) {
	r.delegate.RenderRawMessage(message)
}

func (r FoldableLogsRenderer) printFoldMessage(scopes []string, template string) {
	if scopesCount := len(scopes); scopesCount > 0 {
		lastScope := scopes[scopesCount-1]

		if r.escapeFunc != nil {
			lastScope = r.escapeFunc(lastScope)
		}

		foldingMessage := echelon.NewLogEntryMessage(scopes, echelon.InfoLevel, template, lastScope)
		r.delegate.RenderMessage(foldingMessage)
	}
}
