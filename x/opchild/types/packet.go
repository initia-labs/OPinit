package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/initia-labs/OPinit/x/ophost/types"
)

func DecodePacketData(packetData []byte) (types.AttestorSetUpdatePacketData, error) {
	var data types.AttestorSetUpdatePacketData
	if err := unmarshalProtoJSON(packetData, &data); err != nil {
		return types.AttestorSetUpdatePacketData{}, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	return data, nil
}
