package cli

import (
	"encoding/json"
	"os"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type executeMessages struct {
	// Msgs defines an array of sdk.Msgs proto-JSON-encoded as Anys.
	Messages []json.RawMessage `json:"messages"`
}

// parseExecuteMessages reads and parses the proposal.
func parseExecuteMessages(cdc codec.Codec, path string) ([]sdk.Msg, error) {
	var em executeMessages

	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(contents, &em)
	if err != nil {
		return nil, err
	}

	msgs := make([]sdk.Msg, len(em.Messages))
	for i, anyJSON := range em.Messages {
		var msg sdk.Msg
		err := cdc.UnmarshalInterfaceJSON(anyJSON, &msg)
		if err != nil {
			return nil, err
		}

		msgs[i] = msg
	}

	return msgs, nil
}
