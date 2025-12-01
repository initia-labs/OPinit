package types

const (
	RlyVersion = "v2.6.0"
)

type CosmosProviderConfig struct {
	Key            string  `json:"key"`
	ChainID        string  `json:"chain-id"`
	RPCAddr        string  `json:"rpc-addr"`
	AccountPrefix  string  `json:"account-prefix"`
	KeyringBackend string  `json:"keyring-backend"`
	GasAdjustment  float64 `json:"gas-adjustment"`
	GasPrices      string  `json:"gas-prices"`
	Debug          bool    `json:"debug"`
	Timeout        string  `json:"timeout"`
	OutputFormat   string  `json:"output-format"`
	Broadcast      string  `json:"broadcast-mode"`
}

type ChainConfig struct {
	Type  string               `json:"type"`
	Value CosmosProviderConfig `json:"value"`
}

type PathEnd struct {
	ChainID string `json:"chain-id"`
}

type ChannelFilter struct {
	Rule        string   `json:"rule"`
	ChannelList []string `json:"channel-list"`
}

type Path struct {
	Src    *PathEnd      `json:"src"`
	Dst    *PathEnd      `json:"dst"`
	Filter ChannelFilter `json:"src-channel-filter"`
}
