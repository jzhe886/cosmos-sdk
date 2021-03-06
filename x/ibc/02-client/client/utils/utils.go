package utils

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// QueryAllClientStates returns all the light client states. It _does not_ return
// any merkle proof.
func QueryAllClientStates(cliCtx context.CLIContext, page, limit int) ([]exported.ClientState, int64, error) {
	params := types.NewQueryAllClientsParams(page, limit)
	bz, err := cliCtx.Codec.MarshalJSON(params)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal query params: %w", err)
	}

	route := fmt.Sprintf("custom/%s/%s/%s", "ibc", types.QuerierRoute, types.QueryAllClients)
	res, height, err := cliCtx.QueryWithData(route, bz)
	if err != nil {
		return nil, 0, err
	}

	var clients []exported.ClientState
	err = cliCtx.Codec.UnmarshalJSON(res, &clients)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal light clients: %w", err)
	}
	return clients, height, nil
}

// QueryClientState queries the store to get the light client state and a merkle
// proof.
func QueryClientState(
	cliCtx context.CLIContext, clientID string, prove bool,
) (types.StateResponse, error) {
	req := abci.RequestQuery{
		Path:  "store/ibc/key",
		Data:  ibctypes.KeyClientState(clientID),
		Prove: prove,
	}

	res, err := cliCtx.QueryABCI(req)
	if err != nil {
		return types.StateResponse{}, err
	}

	var clientState exported.ClientState
	if err := cliCtx.Codec.UnmarshalBinaryLengthPrefixed(res.Value, &clientState); err != nil {
		return types.StateResponse{}, err
	}

	clientStateRes := types.NewClientStateResponse(clientID, clientState, res.Proof, res.Height)

	return clientStateRes, nil
}

// QueryConsensusState queries the store to get the consensus state and a merkle
// proof.
func QueryConsensusState(
	cliCtx context.CLIContext, clientID string, height uint64, prove bool,
) (types.ConsensusStateResponse, error) {
	var conStateRes types.ConsensusStateResponse

	req := abci.RequestQuery{
		Path:  "store/ibc/key",
		Data:  ibctypes.KeyConsensusState(clientID, height),
		Prove: prove,
	}

	res, err := cliCtx.QueryABCI(req)
	if err != nil {
		return conStateRes, err
	}

	var cs exported.ConsensusState
	if err := cliCtx.Codec.UnmarshalBinaryLengthPrefixed(res.Value, &cs); err != nil {
		return conStateRes, err
	}

	return types.NewConsensusStateResponse(clientID, cs, res.Proof, res.Height), nil
}

// QueryTendermintHeader takes a client context and returns the appropriate
// tendermint header
func QueryTendermintHeader(cliCtx context.CLIContext) (ibctmtypes.Header, int64, error) {
	node, err := cliCtx.GetNode()
	if err != nil {
		return ibctmtypes.Header{}, 0, err
	}

	info, err := node.ABCIInfo()
	if err != nil {
		return ibctmtypes.Header{}, 0, err
	}

	height := info.Response.LastBlockHeight

	commit, err := node.Commit(&height)
	if err != nil {
		return ibctmtypes.Header{}, 0, err
	}

	validators, err := node.Validators(&height, 0, 10000)
	if err != nil {
		return ibctmtypes.Header{}, 0, err
	}

	header := ibctmtypes.Header{
		SignedHeader: commit.SignedHeader,
		ValidatorSet: tmtypes.NewValidatorSet(validators.Validators),
	}

	return header, height, nil
}

// QueryNodeConsensusState takes a client context and returns the appropriate
// tendermint consensus state
func QueryNodeConsensusState(cliCtx context.CLIContext) (ibctmtypes.ConsensusState, int64, error) {
	node, err := cliCtx.GetNode()
	if err != nil {
		return ibctmtypes.ConsensusState{}, 0, err
	}

	info, err := node.ABCIInfo()
	if err != nil {
		return ibctmtypes.ConsensusState{}, 0, err
	}

	height := info.Response.LastBlockHeight

	commit, err := node.Commit(&height)
	if err != nil {
		return ibctmtypes.ConsensusState{}, 0, err
	}

	validators, err := node.Validators(&height, 0, 10000)
	if err != nil {
		return ibctmtypes.ConsensusState{}, 0, err
	}

	state := ibctmtypes.ConsensusState{
		Timestamp:    commit.Time,
		Root:         commitmenttypes.NewMerkleRoot(commit.AppHash),
		ValidatorSet: tmtypes.NewValidatorSet(validators.Validators),
	}

	return state, height, nil
}
