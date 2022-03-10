package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RegisterInterchainAccount invokes RegisterInterchainAccount which binds a new port for the account owner and initiates the ics27 channel handshake
func (k Keeper) RegisterInterchainAccount(
	ctx sdk.Context,
	owner,
	connectionID string,
) error {
	err := k.icaControllerKeeper.RegisterInterchainAccount(ctx, connectionID, owner)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) InterchainAccountFromAddress(
	ctx sdk.Context,
	owner,
	connectionId string,
) (sdk.AccAddress, error) {

	portID, err := icatypes.NewControllerPortID(owner)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "could not find account: %s", err)
	}

	addr, found := k.icaControllerKeeper.GetInterchainAccountAddress(ctx, connectionId, portID)
	if !found {
		return nil, status.Errorf(codes.NotFound, "no account found for portID %s", portID)
	}

	return sdk.AccAddressFromBech32(addr)
}
