package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"cosmossdk.io/core/appmodule"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/interchain-security/v7/x/ccv/provider/client/cli"
	"github.com/cosmos/interchain-security/v7/x/ccv/provider/keeper"
	"github.com/cosmos/interchain-security/v7/x/ccv/provider/migrations"
	"github.com/cosmos/interchain-security/v7/x/ccv/provider/simulation"
	providertypes "github.com/cosmos/interchain-security/v7/x/ccv/provider/types"
)

var (
	_ module.AppModule           = (*AppModule)(nil)
	_ module.AppModuleBasic      = (*AppModuleBasic)(nil)
	_ module.AppModuleSimulation = (*AppModule)(nil)
	_ module.HasName             = (*AppModule)(nil)
	_ module.HasConsensusVersion = (*AppModule)(nil)
	_ module.HasInvariants       = (*AppModule)(nil)
	_ module.HasServices         = (*AppModule)(nil)
	_ module.HasABCIGenesis      = (*AppModule)(nil)
	_ module.HasABCIEndBlock     = (*AppModule)(nil)
	_ appmodule.AppModule        = (*AppModule)(nil)
	_ appmodule.HasBeginBlocker  = (*AppModule)(nil)
)

// AppModuleBasic is the IBC Provider AppModuleBasic
type AppModuleBasic struct{}

// Name implements AppModuleBasic interface
func (AppModuleBasic) Name() string {
	return providertypes.ModuleName
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// RegisterLegacyAminoCodec implements AppModuleBasic interface
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	providertypes.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers module concrete types into protobuf Any.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	providertypes.RegisterInterfaces(registry)
}

// DefaultGenesis returns default genesis state as raw bytes for the ibc
// provider module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(providertypes.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the ibc provider module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var data providertypes.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", providertypes.ModuleName, err)
	}

	return data.Validate()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the ibc-provider module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	err := providertypes.RegisterQueryHandlerClient(context.Background(), mux, providertypes.NewQueryClient(clientCtx))
	if err != nil {
		// same behavior as in cosmos-sdk
		panic(err)
	}
}

// GetTxCmd implements AppModuleBasic interface
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

// GetQueryCmd implements AppModuleBasic interface
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.NewQueryCmd()
}

// AppModule represents the AppModule for this module
type AppModule struct {
	AppModuleBasic
	keeper     *keeper.Keeper
	paramSpace paramtypes.Subspace
	storeKey   storetypes.StoreKey
}

// NewAppModule creates a new provider module
func NewAppModule(k *keeper.Keeper, paramSpace paramtypes.Subspace, storeKey storetypes.StoreKey) AppModule {
	return AppModule{
		keeper:     k,
		paramSpace: paramSpace,
		storeKey:   storeKey,
	}
}

// RegisterInvariants implements the AppModule interface
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	keeper.RegisterInvariants(ir, am.keeper)
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	providertypes.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	providertypes.RegisterQueryServer(cfg.QueryServer(), am.keeper)

	migrator := migrations.NewMigrator(*am.keeper, am.paramSpace, am.storeKey)
	if err := cfg.RegisterMigration(providertypes.ModuleName, 2, migrator.Migrate2to3); err != nil {
		panic(fmt.Sprintf("failed to register migrator for %s: %s -- from 2 -> 3", providertypes.ModuleName, err))
	}
	if err := cfg.RegisterMigration(providertypes.ModuleName, 3, migrator.Migrate3to4); err != nil {
		panic(fmt.Sprintf("failed to register migrator for %s: %s -- from 3 -> 4", providertypes.ModuleName, err))
	}
	if err := cfg.RegisterMigration(providertypes.ModuleName, 4, migrator.Migrate4to5); err != nil {
		panic(fmt.Sprintf("failed to register migrator for %s: %s -- from 4 -> 5", providertypes.ModuleName, err))
	}
	if err := cfg.RegisterMigration(providertypes.ModuleName, 5, migrator.Migrate5to6); err != nil {
		panic(fmt.Sprintf("failed to register migrator for %s: %s -- from 5 -> 6", providertypes.ModuleName, err))
	}
	if err := cfg.RegisterMigration(providertypes.ModuleName, 6, migrator.Migrate6to7); err != nil {
		panic(fmt.Sprintf("failed to register migrator for %s: %s -- from 6 -> 7", providertypes.ModuleName, err))
	}
	if err := cfg.RegisterMigration(providertypes.ModuleName, 7, migrator.Migrate7to8); err != nil {
		panic(fmt.Sprintf("failed to register migrator for %s: %s -- from 7 -> 8", providertypes.ModuleName, err))
	}
}

// InitGenesis performs genesis initialization for the provider module. It returns validator updates
// by selecting the first MaxProviderConsensusValidators from the staking module's validator set.
// Note: This method along with ValidateGenesis satisfies the CCV spec:
// https://github.com/cosmos/ibc/blob/main/spec/app/ics-028-cross-chain-validation/methods.md#ccv-pcf-initg1
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState providertypes.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)

	return am.keeper.InitGenesis(ctx, &genesisState)
}

// ExportGenesis returns the exported genesis state as raw bytes for the provider
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := am.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(gs)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 8 }

// BeginBlock implements the AppModule interface
func (am AppModule) BeginBlock(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Create clients to consumer chains that are due to be spawned
	if err := am.keeper.BeginBlockLaunchConsumers(sdkCtx); err != nil {
		return err
	}
	// Stop and remove state for any consumer chains that are due to be stopped
	if err := am.keeper.BeginBlockRemoveConsumers(sdkCtx); err != nil {
		return err
	}
	// Update the infraction parameters for consumer chains that are scheduled for an update
	if err := am.keeper.BeginBlockUpdateInfractionParameters(sdkCtx); err != nil {
		return err
	}
	// Check for replenishing slash meter before any slash packets are processed for this block
	am.keeper.BeginBlockCIS(sdkCtx)
	// BeginBlock logic needed for the  Reward Distribution sub-protocol
	am.keeper.BeginBlockRD(sdkCtx)

	return nil
}

// EndBlock implements the AppModule interface
func (am AppModule) EndBlock(ctx context.Context) ([]abci.ValidatorUpdate, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// EndBlock logic needed for the Consumer Initiated Slashing sub-protocol.
	// Important: EndBlockCIS must be called before EndBlockVSU
	am.keeper.EndBlockCIS(sdkCtx)
	// EndBlock logic needed for the Validator Set Update sub-protocol
	return am.keeper.EndBlockVSU(sdkCtx)
}

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the transfer module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

// RegisterStoreDecoder registers a decoder for provider module's types
func (am AppModule) RegisterStoreDecoder(sdr simtypes.StoreDecoderRegistry) {
}

// WeightedOperations returns the all the provider module operations with their respective weights.
func (am AppModule) WeightedOperations(_ module.SimulationState) []simtypes.WeightedOperation {
	return nil
}
