package agents

import (
	"agentflow/internal/agents/tools"
	"context"
	"fmt"

	"github.com/nlpodyssey/openai-agents-go/agents"
	"github.com/nlpodyssey/openai-agents-go/modelsettings"
	"github.com/openai/openai-go/v2"
)

type Agent struct {
	Agent *agents.Agent
}

func (a *Agent) RunInputs(ctx context.Context, prompts []agents.TResponseInputItem) (string, error) {
	// loop through prompts and print them
	for _, prompt := range prompts {
		fmt.Println(prompt.OfMessage.Content.OfString.String())
	}
	result, err := agents.RunInputs(ctx, a.Agent, prompts)
	if err != nil {
		return "", err
	}
	return fmt.Sprint(result.FinalOutput), nil
}

var (
	PO *Agent
	SA *Agent
	LD *Agent
	LQ *Agent
)

func init() {
	PO = newAgent("Product Owner", "", "gpt-5")
	SA = newAgent("Solution Architect", "", "gpt-5")
	LD = newAgent("Lead Developer", "", "gpt-5")
	LQ = newAgent("Lead QA", "", "gpt-5")
}

func newAgent(role, instructions, model string) *Agent {
	return &Agent{
		Agent: agents.New(role).
			WithInstructions(instructions).
			WithModel(model).
			WithTools(
				tools.FileCreatorTool,
				tools.FileReaderTool,
			).
			WithModelSettings(modelsettings.ModelSettings{
				Temperature: openai.Float(1.0),
			}),
	}
}

func UserMessage(message string) agents.TResponseInputItem {
	return agents.UserMessage(message)
}

func InputList(items ...any) []agents.TResponseInputItem {
	return agents.InputList(items...)
}

type TResponseInputItem = agents.TResponseInputItem

func SystemMessage(message string) agents.TResponseInputItem {
	return agents.SystemMessage(message)
}
