package cli

import "github.com/initia-labs/OPinit/x/ophost/types"

// BridgeConfig defines the set of bridge config.
// NOTE: it is a modified BridgeConfig from x/ophost/types/types.go to make unmarshal easier
type BridgeCliConfig struct {
	Challengers           []string        `json:"challengers"`
	Proposer              string          `json:"proposer"`
	SubmissionInterval    string          `json:"submission_interval"`
	FinalizationPeriod    string          `json:"finalization_period"`
	SubmissionStartHeight string          `json:"submission_start_height"`
	Metadata              string          `json:"metadata"`
	BatchInfo             types.BatchInfo `json:"batch_info"`
}

// MsgFinalizeTokenWithdrawal is a message to remove a validator from designated list
//
// NOTE: it is a modified MsgFinalizeTokenWithdrawal from x/ophost/types/tx.pb.go to make unmarshal easier
type MsgFinalizeTokenWithdrawal struct {
	BridgeId         uint64   `protobuf:"varint,2,opt,name=bridge_id,json=bridgeId,proto3" json:"bridge_id,omitempty" yaml:"bridge_id"`
	OutputIndex      uint64   `protobuf:"varint,3,opt,name=output_index,json=outputIndex,proto3" json:"output_index,omitempty" yaml:"output_index"`
	WithdrawalProofs [][]byte `protobuf:"bytes,4,rep,name=withdrawal_proofs,json=withdrawalProofs,proto3" json:"withdrawal_proofs,omitempty"`
	Sender           string   `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty" yaml:"sender"`
	// no receiver here
	// Receiver string `protobuf:"bytes,5,opt,name=receiver,proto3" json:"receiver,omitempty" yaml:"receiver"`
	Sequence uint64 `protobuf:"varint,6,opt,name=sequence,proto3" json:"sequence,omitempty" yaml:"sequence"`
	Amount   string `protobuf:"bytes,7,opt,name=amount,proto3" json:"amount" yaml:"amount"`
	// version of the output root
	Version         []byte `protobuf:"bytes,8,opt,name=version,proto3" json:"version,omitempty" yaml:"version"`
	StorageRoot     []byte `protobuf:"bytes,10,opt,name=storage_root,json=storageRoot,proto3" json:"storage_root,omitempty" yaml:"storage_root"`
	LatestBlockHash []byte `protobuf:"bytes,11,opt,name=latest_block_hash,json=latestBlockHash,proto3" json:"latest_block_hash,omitempty" yaml:"latest_block_hash"`
}
