package cli

import "github.com/initia-labs/OPinit/x/ophost/types"

// BridgeConfig defines the set of bridge config.
//
// NOTE: it is a modified BridgeConfig from x/ophost/types/types.pb.go to make unmarshal easier
type BridgeConfig struct {
	// The address of the challenger.
	Challenger string `protobuf:"bytes,1,opt,name=challenger,proto3" json:"challenger,omitempty"`
	// The address of the proposer.
	Proposer string `protobuf:"bytes,2,opt,name=proposer,proto3" json:"proposer,omitempty"`
	// The time interval at which checkpoints must be submitted.
	// NOTE: this param is currently not used, but will be used for challenge in future.
	SubmissionInterval string `protobuf:"bytes,3,opt,name=submission_interval,json=submissionInterval,proto3,stdduration" json:"submission_interval,omitempty"`
	// The minium time duration that must elapse before a withdrawal can be finalized.
	FinalizationPeriod string `protobuf:"bytes,4,opt,name=finalization_period,json=finalizationPeriod,proto3,stdduration" json:"finalization_period,omitempty"`
	// The time of the first l2 block recorded.
	// NOTE: this param is currently not used, but will be used for challenge in future.
	SubmissionStartTime string `protobuf:"bytes,5,opt,name=submission_start_time,json=submissionStartTime,proto3,stdtime" json:"submission_start_time"`
	// Normally it is IBC channelID for permissioned IBC relayer.
	Metadata string `protobuf:"bytes,6,opt,name=metadata,proto3" json:"metadata,omitempty"`
	// BatchInfo is the batch information for the bridge.
	BatchInfo types.BatchInfo `json:"batch_info"`
}

// MsgFinalizeTokenWithdrawal is a message to remove a validator from designated list
//
// NOTE: it is a modified MsgFinalizeTokenWithdrawal from x/ophost/types/tx.pb.go to make unmarshal easier
type MsgFinalizeTokenWithdrawal struct {
	BridgeId         uint64   `protobuf:"varint,2,opt,name=bridge_id,json=bridgeId,proto3" json:"bridge_id,omitempty" yaml:"bridge_id"`
	OutputIndex      uint64   `protobuf:"varint,3,opt,name=output_index,json=outputIndex,proto3" json:"output_index,omitempty" yaml:"output_index"`
	WithdrawalProofs []string `protobuf:"bytes,4,rep,name=withdrawal_proofs,json=withdrawalProofs,proto3" json:"withdrawal_proofs,omitempty"`
	// no sender here
	//Sender           string   `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty" yaml:"sender"`
	Receiver string `protobuf:"bytes,5,opt,name=receiver,proto3" json:"receiver,omitempty" yaml:"receiver"`
	Sequence uint64 `protobuf:"varint,6,opt,name=sequence,proto3" json:"sequence,omitempty" yaml:"sequence"`
	Amount   string `protobuf:"bytes,7,opt,name=amount,proto3" json:"amount" yaml:"amount"`
	// version of the output root
	Version         string `protobuf:"bytes,8,opt,name=version,proto3" json:"version,omitempty" yaml:"version"`
	StateRoot       string `protobuf:"bytes,9,opt,name=state_root,json=stateRoot,proto3" json:"state_root,omitempty" yaml:"state_root"`
	StorageRoot     string `protobuf:"bytes,10,opt,name=storage_root,json=storageRoot,proto3" json:"storage_root,omitempty" yaml:"storage_root"`
	LatestBlockHash string `protobuf:"bytes,11,opt,name=latest_block_hash,json=latestBlockHash,proto3" json:"latest_block_hash,omitempty" yaml:"latest_block_hash"`
}
