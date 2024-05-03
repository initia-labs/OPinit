package utils

import (
	abci "github.com/cometbft/cometbft/abci/types"
)

// FindTxEventsByKey returns the value of the first attribute with the given key.
func FindTxEventsByKey(key string, events []abci.Event) (string, bool) {
	for _, ev := range events {
		for _, attr := range ev.Attributes {
			if attr.Key == key {
				return attr.Value, true
			}
		}
	}

	return "", false
}
