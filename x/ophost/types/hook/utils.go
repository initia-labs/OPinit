package hook

import (
	"encoding/json"
	"strings"

	errorsmod "cosmossdk.io/errors"

	"github.com/initia-labs/OPinit/x/ophost/types"
)

const permsMetadataKey = "perm_channels"

type PermsMetadata struct {
	PermChannels []PortChannelID `json:"perm_channels"`
}

type PortChannelID struct {
	PortID    string `json:"port_id,omitempty"`
	ChannelID string `json:"channel_id,omitempty"`
}

func hasPermChannels(metadata []byte) (hasPerms bool, data PermsMetadata) {
	if !jsonStringHasKey(string(metadata), permsMetadataKey) {
		return false, data
	}

	decoder := json.NewDecoder(strings.NewReader(string(metadata)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&data); err != nil {
		return false, data
	}
	return true, data
}

// jsonStringHasKey parses the metadata string as a json object and checks if it contains the key.
func jsonStringHasKey(metadata, key string) bool {
	if len(metadata) == 0 {
		return false
	}

	jsonObject := make(map[string]interface{})
	err := json.Unmarshal([]byte(metadata), &jsonObject)
	if err != nil {
		return false
	}

	_, ok := jsonObject[key]
	return ok
}

// GetOpinitChannelID extracts the opinit channel ID from the bridge metadata.
// Returns the channel ID for the opinit port, or an error if not found.
func GetOpinitChannelID(metadata []byte) (string, error) {
	var permsMetadata PermsMetadata
	if err := json.Unmarshal(metadata, &permsMetadata); err != nil {
		return "", errorsmod.Wrap(err, "failed to unmarshal metadata")
	}

	for _, permChannel := range permsMetadata.PermChannels {
		if permChannel.PortID == types.PortID {
			if permChannel.ChannelID == "" {
				return "", errorsmod.Wrap(types.ErrInvalidBridgeMetadata, "opinit channel ID is empty")
			}
			return permChannel.ChannelID, nil
		}
	}

	return "", errorsmod.Wrap(types.ErrInvalidBridgeMetadata, "opinit channel not found in metadata")
}
