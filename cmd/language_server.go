package cmd

import (
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/tliron/commonlog"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/tliron/glsp/server"
	"github.com/urfave/cli/v2"

	_ "github.com/tliron/commonlog/simple"
)

func LanguageServer() {
	commonlog.Configure(2, nil)

	server.NewServer(
		&protocol.Handler{
			Initialize:  initialize,
			Initialized: initialized,
			Shutdown:    shutdown,
			SetTrace:    setTrace,
			// TextDocumentSemanticTokensFull: textDocumentSemanticTokensFull,
			// TextDocumentDidChange: textDocumentDidChange,
		},
		"jet",
		false,
	).RunStdio()
}

func actionLanguageServer(*cli.Context) error {
	LanguageServer()
	return nil
}

func ptr[T any](value T) *T { return &value }

func initialize(ctx *glsp.Context, params *protocol.InitializeParams) (any, error) {
	result := protocol.InitializeResult{
		ServerInfo: &protocol.InitializeResultServerInfo{
			Name:    "Jet language server",
			Version: ptr("0.0.1"),
		},
		Capabilities: protocol.ServerCapabilities{
			// SemanticTokensProvider: &protocol.SemanticTokensOptions{
			// 	Legend: protocol.SemanticTokensLegend{
			// 		TokenTypes: []string{
			// 			string(protocol.SemanticTokenTypeNamespace),
			// 			string(protocol.SemanticTokenTypeType),
			// 			string(protocol.SemanticTokenTypeClass),
			// 			string(protocol.SemanticTokenTypeEnum),
			// 			string(protocol.SemanticTokenTypeInterface),
			// 			string(protocol.SemanticTokenTypeStruct),
			// 			string(protocol.SemanticTokenTypeTypeParameter),
			// 			string(protocol.SemanticTokenTypeParameter),
			// 			string(protocol.SemanticTokenTypeVariable),
			// 			string(protocol.SemanticTokenTypeProperty),
			// 			string(protocol.SemanticTokenTypeEnumMember),
			// 			string(protocol.SemanticTokenTypeEvent),
			// 			string(protocol.SemanticTokenTypeFunction),
			// 			string(protocol.SemanticTokenTypeMethod),
			// 			string(protocol.SemanticTokenTypeMacro),
			// 			string(protocol.SemanticTokenTypeKeyword),
			// 			string(protocol.SemanticTokenTypeModifier),
			// 			string(protocol.SemanticTokenTypeComment),
			// 			string(protocol.SemanticTokenTypeString),
			// 			string(protocol.SemanticTokenTypeNumber),
			// 			string(protocol.SemanticTokenTypeRegexp),
			// 			string(protocol.SemanticTokenTypeOperator),
			// 		},
			// 		TokenModifiers: []string{
			// 			string(protocol.SemanticTokenModifierDeclaration),
			// 			string(protocol.SemanticTokenModifierDefinition),
			// 			string(protocol.SemanticTokenModifierReadonly),
			// 			string(protocol.SemanticTokenModifierStatic),
			// 			string(protocol.SemanticTokenModifierDeprecated),
			// 			string(protocol.SemanticTokenModifierAbstract),
			// 			string(protocol.SemanticTokenModifierAsync),
			// 			string(protocol.SemanticTokenModifierModification),
			// 			string(protocol.SemanticTokenModifierDocumentation),
			// 			string(protocol.SemanticTokenModifierDefaultLibrary),
			// 		},
			// 	},
			// 	Full: ptr(true),
			// },
		},
	}
	spew.Fprintf(os.Stderr, "%s: %s\n", protocol.MethodInitialize, spew.Sdump(params))
	spew.Fprintf(os.Stderr, "response: %s\n", spew.Sdump(result))
	return result, nil
}

func initialized(ctx *glsp.Context, params *protocol.InitializedParams) error {
	spew.Fprintf(os.Stderr, "%s: %s\n", protocol.MethodInitialized, spew.Sdump(params))
	return nil
}

func shutdown(context *glsp.Context) error {
	spew.Fprintf(os.Stderr, "%s\n", protocol.MethodShutdown)
	protocol.SetTraceValue(protocol.TraceValueOff)
	return nil
}

func setTrace(ctx *glsp.Context, params *protocol.SetTraceParams) error {
	spew.Fprintf(os.Stderr, "%s: %s\n", protocol.MethodSetTrace, spew.Sdump(params))
	protocol.SetTraceValue(params.Value)
	return nil
}

func textDocumentSemanticTokensFull(
	ctx *glsp.Context,
	params *protocol.SemanticTokensParams,
) (*protocol.SemanticTokens, error) {
	result := &protocol.SemanticTokens{
		Data: []protocol.UInteger{
			// 0, 4, 1, 8, 3,
			// 0, 8, 1, 8, 0,
		},
	}
	spew.Fprintf(os.Stderr, "%s: %s\n", protocol.MethodTextDocumentSemanticTokensFull, spew.Sdump(params))
	spew.Fprintf(os.Stderr, "response: %s\n", spew.Sdump(result))
	return result, nil
}

func textDocumentDidChange(
	ctx *glsp.Context,
	params *protocol.DidChangeTextDocumentParams,
) error {
	spew.Fprintf(os.Stderr, "%s: %s\n", protocol.MethodTextDocumentDidChange, spew.Sdump(params))
	return nil
}
