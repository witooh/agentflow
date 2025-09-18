package agents

import (
	"context"
	"fmt"

	"github.com/nlpodyssey/openai-agents-go/agents"
	"github.com/nlpodyssey/openai-agents-go/modelsettings"
	"github.com/openai/openai-go/v2"
)

type Agent struct {
	Agent *agents.Agent
}

func (a *Agent) Run(ctx context.Context, prompt string) (string, error) {
	result, err := agents.Run(ctx, a.Agent, prompt)
	if err != nil {
		return "", err
	}
	fmt.Println(result)
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
			WithModelSettings(modelsettings.ModelSettings{
				Temperature: openai.Float(1.0),
			}),
	}
}
