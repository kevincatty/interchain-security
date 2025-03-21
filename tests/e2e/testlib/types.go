package e2e

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"reflect"
	"time"

	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"

	"github.com/cosmos/interchain-security/v7/x/ccv/provider/types"
)

type (
	ChainID     string
	ConsumerID  string
	ValidatorID string
)

type AssignConsumerPubKeyAction struct {
	Chain          ChainID
	Validator      ValidatorID
	ConsumerPubkey string
	// ReconfigureNode will change keys the node uses and restart
	ReconfigureNode bool
	// executing the action should raise an error
	ExpectError   bool
	ExpectedError string
}

type SubmitConsumerAdditionProposalAction struct {
	PreCCV              bool
	Chain               ChainID
	From                ValidatorID
	Deposit             uint
	ConsumerChain       ChainID
	SpawnTime           uint
	InitialHeight       clienttypes.Height
	DistributionChannel string
	TopN                uint32
	ValidatorsPowerCap  uint32
	ValidatorSetCap     uint32
	Allowlist           []string
	Denylist            []string
	MinStake            uint64
	AllowInactiveVals   bool
	Prioritylist        []string
}

type SubmitConsumerRemovalProposalAction struct {
	Chain          ChainID
	From           ValidatorID
	Deposit        uint
	ConsumerChain  ChainID
	StopTimeOffset time.Duration // offset from time.Now()
}

type StartChainAction struct {
	Chain      ChainID
	Validators []StartChainValidator
	// Genesis changes specific to this action, appended to genesis changes defined in chain config
	GenesisChanges string
	IsConsumer     bool
}

type StartChainValidator struct {
	Id         ValidatorID
	Allocation uint
	Stake      uint
}

type StartConsumerChainAction struct {
	ConsumerChain  ChainID
	ProviderChain  ChainID
	Validators     []StartChainValidator
	GenesisChanges string
}

type DelegateTokensAction struct {
	Chain  ChainID
	From   ValidatorID
	To     ValidatorID
	Amount uint
}

type UnbondTokensAction struct {
	Chain      ChainID
	Sender     ValidatorID
	UnbondFrom ValidatorID
	Amount     uint
}

type ChainCommands interface {
	// State commands - functions use by test driver to get state information
	GetBlockHeight(chain ChainID) uint
	GetBalance(chain ChainID, validator ValidatorID) uint
	GetConsumerChains(chain ChainID) map[ChainID]bool
	GetConsumerAddress(consumerChain ChainID, validator ValidatorID) string
	GetClientFrozenHeight(chain ChainID, clientID string) (RevisionNumber, RevisionHeight uint64)
	GetHasToValidate(validator ValidatorID) []ChainID
	GetIBCTransferParams(chain ChainID) IBCTransferParams
	GetProposal(chain ChainID, proposal uint) Proposal
	GetParam(chain ChainID, param Param) string
	GetProviderAddressFromConsumer(consumerChain ChainID, validator ValidatorID) string
	GetReward(chain ChainID, validator ValidatorID, blockHeight uint, denom string) float64
	GetRegisteredConsumerRewardDenoms(chain ChainID) []string
	GetSlashMeter() int64
	GetPendingPacketQueueSize(chain ChainID) uint
	GetProposedConsumerChains(chain ChainID) []string
	GetQueryNode(chain ChainID) string
	GetQueryNodeRPCAddress(chain ChainID) string
	GetTrustedHeight(chain ChainID, clientID string, index int) (uint64, uint64)
	GetValPower(chain ChainID, validator ValidatorID) uint
	GetValStakedTokens(chain ChainID, validatorAddress string) uint
	GetQueryNodeIP(chain ChainID) string
	GetInflationRate(chain ChainID) float64
	GetConsumerCommissionRate(chain ChainID, validator ValidatorID) float64
	QueryTransaction(chain ChainID, txhash string) ([]byte, error)

	CreateConsumer(providerChain, consumerChain ChainID, validator ValidatorID, metadata types.ConsumerMetadata, initParams *types.ConsumerInitializationParameters, powerShapingParams *types.PowerShapingParameters) ([]byte, error)
	UpdateConsumer(providerChain ChainID, validator ValidatorID, update types.MsgUpdateConsumer, verbose bool) ([]byte, error)
	SubmitGovProposal(chain ChainID, from ValidatorID, command, proposal string, verbose bool) ([]byte, error)
	AssignConsumerPubKey(identifier, pubKey string, from ValidatorID, gas, home, node string, verbose bool) ([]byte, error)
}

type ActionCommands interface {
	SubmitConsumerAdditionProposal(action SubmitConsumerAdditionProposalAction, verbose bool)
	AssignConsumerPubKey(action AssignConsumerPubKeyAction, verbose bool)
	StartChain(action StartChainAction, verbose bool)
	StartConsumerChain(action StartConsumerChainAction, verbose bool)
	SubmitConsumerRemovalProposal(action SubmitConsumerRemovalProposalAction, verbose bool)
	DelegateTokens(action DelegateTokensAction, verbose bool)
	UnbondTokens(action UnbondTokensAction, verbose bool)
}
type ChainIF interface {
	ActionCommands
}

type ActionHandler func(action interface{}, verbose bool) error

// TODO: replace ExecutionTarget with new TargetDriver interface
type PlatformDriver interface {
	ExecCommand(name string, arg ...string) *exec.Cmd
	// ExecDetachedCommand: when executed the command will be run in the background and call will return immediately
	ExecDetachedCommand(name string, args ...string) *exec.Cmd
	GetTestScriptPath(isConsumer bool, script string) string
	UseCometMock() bool
}
type TargetDriver interface {
	// ChainCommands
	ChainCommands
	PlatformDriver
}

// TODO: this should not be here. mv 'Now' to a better suited type here and then move ContainerConfig back
type ContainerConfig struct {
	ContainerName string
	CcvVersion    string
	Now           time.Time
}

// Attributes that are unique to a validator. Allows us to map (part of)
// the set of strings defined above to a set of viable validators
type ValidatorConfig struct {
	// Seed phrase to generate a secp256k1 key used by the validator on the provider
	Mnemonic string
	// Validator account address on provider marshaled to string using Bech32
	// with Bech32Prefix = ProviderAccountPrefix
	DelAddress string
	// Validator account address on provider marshaled to string using Bech32
	// with Bech32Prefix = ConsumerAccountPrefix
	DelAddressOnConsumer string
	// Validator operator address on provider marshaled to string using Bech32
	// with Bech32Prefix = ProviderAccountPrefix
	ValoperAddress string
	// Validator operator address on provider marshaled to string using Bech32
	// with Bech32Prefix = ConsumerAccountPrefix
	ValoperAddressOnConsumer string
	// Validator consensus address on provider marshaled to string using Bech32
	// with Bech32Prefix = ProviderAccountPrefix. It matches the PrivValidatorKey below.
	ValconsAddress string
	// Validator consensus address on provider marshaled to string using Bech32
	// with Bech32Prefix = ConsumerAccountPrefix.
	ValconsAddressOnConsumer string
	// Key used for consensus on provider
	PrivValidatorKey string
	NodeKey          string
	// Must be an integer greater than 0 and less than 253
	IpSuffix string

	// consumer chain key assignment data
	// keys are used on a new node

	// Seed phrase to generate a secp256k1 key used by the validator on the consumer
	ConsumerMnemonic string
	// Validator account address on consumer marshaled to string using Bech32
	// with Bech32Prefix = ConsumerAccountPrefix
	ConsumerDelAddress string
	// Validator account address on consumer marshaled to string using Bech32
	// with Bech32Prefix = ProviderAccountPrefix
	ConsumerDelAddressOnProvider string
	// Validator operator address on consumer marshaled to string using Bech32
	// with Bech32Prefix = ConsumerAccountPrefix
	ConsumerValoperAddress string
	// Validator operator address on consumer marshaled to string using Bech32
	// with Bech32Prefix = ProviderAccountPrefix
	ConsumerValoperAddressOnProvider string
	// Validator consensus address on consumer marshaled to string using Bech32
	// with Bech32Prefix = ConsumerAccountPrefix. It matches the PrivValidatorKey below.
	ConsumerValconsAddress string
	// Validator consensus address on consumer marshaled to string using Bech32
	// with Bech32Prefix = ProviderAccountPrefix.
	ConsumerValconsAddressOnProvider string
	ConsumerValPubKey                string
	// Key used for consensus on consumer
	ConsumerPrivValidatorKey string
	ConsumerNodeKey          string
	UseConsumerKey           bool // if true the validator node will start with consumer key
}

// Attributes that are unique to a chain. Allows us to map (part of)
// the set of strings defined above to a set of viable chains
type ChainConfig struct {
	ChainId    ChainID
	ConsumerId ConsumerID
	// The account prefix configured on the chain. For example, on the Hub, this is "cosmos"
	AccountPrefix string
	// Must be unique per chain
	IpPrefix       string
	VotingWaitTime uint
	// Any transformations to apply to the genesis file of all chains instantiated with this chain config, as a jq string.
	// Example: ".app_state.gov.params.voting_period = \"5s\" | .app_state.slashing.params.signed_blocks_window = \"2\" | .app_state.slashing.params.min_signed_per_window = \"0.500000000000000000\""
	GenesisChanges string
	BinaryName     string

	// binary to use after upgrade height
	UpgradeBinary string
}

type (
	// to have a ChainState object that does not have the overridden Marshal/Unmarshal method
	ChainStateCopy ChainState

	// duplicated from the ChainState with a minor change to the Proposals field
	ChainStateWithProposalTypes struct {
		ChainStateCopy
		Proposals *map[uint]ProposalAndType // the only thing changed from the real ChainState
	}
)

// stores a proposal as a raw json, together with its type
type ProposalAndType struct {
	RawProposal json.RawMessage
	Type        string
}
type ChainState struct {
	ValBalances                    *map[ValidatorID]uint
	Proposals                      *map[uint]Proposal
	ProposedConsumerChains         *[]string
	ValPowers                      *map[ValidatorID]uint
	StakedTokens                   *map[ValidatorID]uint
	IBCTransferParams              *IBCTransferParams
	Params                         *[]Param
	Rewards                        *Rewards
	ConsumerChains                 *map[ChainID]bool
	AssignedKeys                   *map[ValidatorID]string
	ProviderKeys                   *map[ValidatorID]string // validatorID: validator provider key
	ConsumerPendingPacketQueueSize *uint                   // Only relevant to consumer chains
	RegisteredConsumerRewardDenoms *[]string
	ClientsFrozenHeights           *map[string]clienttypes.Height
	HasToValidate                  *map[ValidatorID][]ChainID // only relevant to provider chain
	InflationRateChange            *int                       // whether the inflation rate between two blocks changes negatively (any negative number), is equal (0), or changes positively (any positive number)
	ConsumerCommissionRates        *map[ValidatorID]float64
}

// custom marshal and unmarshal functions for the chainstate that convert proposals to/from the auxiliary type with type info

// MarshalJSON transforms the ChainState into a ChainStateWithProposalTypes by adding type info to the proposals
func (c ChainState) MarshalJSON() ([]byte, error) {
	chainStateCopy := ChainStateCopy(c)
	chainStateWithProposalTypes := ChainStateWithProposalTypes{chainStateCopy, nil}
	if c.Proposals != nil {
		proposalsWithTypes := make(map[uint]ProposalAndType)
		for k, v := range *c.Proposals {
			rawMessage, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}
			proposalsWithTypes[k] = ProposalAndType{rawMessage, reflect.TypeOf(v).String()}
		}
		chainStateWithProposalTypes.Proposals = &proposalsWithTypes
	}
	return json.Marshal(chainStateWithProposalTypes)
}

// UnmarshalJSON unmarshals the ChainStateWithProposalTypes into a ChainState by removing the type info from the proposals and getting back standard proposals
func (c *ChainState) UnmarshalJSON(data []byte) error {
	chainStateWithProposalTypes := ChainStateWithProposalTypes{}
	err := json.Unmarshal(data, &chainStateWithProposalTypes)
	if err != nil {
		return err
	}

	chainState := ChainState(chainStateWithProposalTypes.ChainStateCopy)
	*c = chainState

	if chainStateWithProposalTypes.Proposals != nil {
		proposals := make(map[uint]Proposal)
		for k, v := range *chainStateWithProposalTypes.Proposals {
			proposal, err := UnmarshalProposalWithType(v.RawProposal, v.Type)
			if err != nil {
				return err
			}
			proposals[k] = proposal
		}
		c.Proposals = &proposals
	}
	return nil
}

// UnmarshalProposalWithType takes a JSON object and a proposal type and marshals into an object of the corresponding proposal.
func UnmarshalProposalWithType(inputMap json.RawMessage, proposalType string) (Proposal, error) {
	var err error
	switch proposalType {
	case "main.TextProposal":
		prop := TextProposal{}
		err := json.Unmarshal(inputMap, &prop)
		if err == nil {
			return prop, nil
		}
	case "main.ConsumerAdditionProposal":
		prop := ConsumerAdditionProposal{}
		err := json.Unmarshal(inputMap, &prop)
		if err == nil {
			return prop, nil
		}
	case "main.UpgradeProposal":
		prop := UpgradeProposal{}
		err := json.Unmarshal(inputMap, &prop)
		if err == nil {
			return prop, nil
		}
	case "main.ConsumerRemovalProposal":
		prop := ConsumerRemovalProposal{}
		err := json.Unmarshal(inputMap, &prop)
		if err == nil {
			return prop, nil
		}
	case "main.IBCTransferParamsProposal":
		prop := IBCTransferParamsProposal{}
		err := json.Unmarshal(inputMap, &prop)
		if err == nil {
			return prop, nil
		}
	default:
		return nil, fmt.Errorf("%s is not a known proposal type", proposalType)
	}

	return nil, err
}

type Proposal interface {
	isProposal()
}
type TextProposal struct {
	Title       string
	Description string
	Deposit     uint
	Status      string
}

func (p TextProposal) isProposal() {}

type IBCTransferParamsProposal struct {
	Title   string
	Deposit uint
	Status  string
	Params  IBCTransferParams
}

func (ibct IBCTransferParamsProposal) isProposal() {}

type ConsumerAdditionProposal struct {
	Deposit       uint
	Chain         ChainID
	SpawnTime     int
	InitialHeight clienttypes.Height
	Status        string
}

type UpgradeProposal struct {
	Title         string
	Description   string
	UpgradeHeight uint64
	Type          string
	Deposit       uint
	Status        string
}

func (p UpgradeProposal) isProposal() {}

func (p ConsumerAdditionProposal) isProposal() {}

type ConsumerRemovalProposal struct {
	Deposit uint
	Chain   ChainID
	Status  string
}

func (p ConsumerRemovalProposal) isProposal() {}

type ConsumerModificationProposal struct {
	Deposit uint
	Chain   ChainID
	Status  string
}

func (p ConsumerModificationProposal) isProposal() {}

type Rewards struct {
	IsRewarded map[ValidatorID]bool
	// if true it will calculate if the validator/delegator is rewarded between 2 successive blocks,
	// otherwise it will calculate if it received any rewards since the 1st block
	IsIncrementalReward bool
	// The reward denom to be checked. This can be either the native "stake" denom or
	// a denom from other chains (e.g. if provider received rewards from a consumer chain)
	Denom string
}

type ParamsProposal struct {
	Deposit  uint
	Status   string
	Subspace string
	Key      string
	Value    string
}

func (p ParamsProposal) isProposal() {}

type Param struct {
	Subspace string
	Key      string
	Value    string
}

type IBCTransferParams struct {
	SendEnabled    bool `json:"send_enabled"`
	ReceiveEnabled bool `json:"receive_enabled"`
}
