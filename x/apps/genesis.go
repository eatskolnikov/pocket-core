package pos

import (
	"fmt"
	"github.com/pokt-network/pocket-core/x/apps/keeper"
	"github.com/pokt-network/pocket-core/x/apps/types"
	sdk "github.com/pokt-network/posmint/types"
)

// InitGenesis sets up the module based on the genesis state
// First TM block is at height 1, so state updates applied from
// genesis.json are in block 0.
func InitGenesis(ctx sdk.Context, keeper keeper.Keeper, supplyKeeper types.SupplyKeeper, posKeeper types.PosKeeper, data types.GenesisState) {
	stakedTokens := sdk.ZeroInt()
	ctx = ctx.WithBlockHeight(1 - sdk.ValidatorUpdateDelay)
	// set the parameters from the data
	keeper.SetParams(ctx, data.Params)
	for _, application := range data.Applications {
		if application.IsUnstaked() || application.IsUnstaking() {
			panic(fmt.Sprintf("%v the applications must be staked at genesis", application))
		}
		// Call the registration hook if not exported
		if !data.Exported {
			keeper.BeforeApplicationRegistered(ctx, application.Address)
		}
		// calculate relays
		application.MaxRelays = keeper.CalculateAppRelays(ctx, application)
		// set the applications from the data
		keeper.SetApplication(ctx, application)
		keeper.SetStakedApplication(ctx, application)
		if !data.Exported {
			keeper.AfterApplicationRegistered(ctx, application.Address)
		}
		if application.IsStaked() {
			stakedTokens = stakedTokens.Add(application.GetTokens())
		}
	}
	stakedCoins := sdk.NewCoins(sdk.NewCoin(posKeeper.StakeDenom(ctx), stakedTokens))
	// check if the staked pool accounts exists
	stakedPool := keeper.GetStakedPool(ctx)
	if stakedPool == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.StakedPoolName))
	}
	// add coins if not provided on genesis
	if stakedPool.GetCoins().IsZero() {
		if err := stakedPool.SetCoins(stakedCoins); err != nil {
			panic(err)
		}
		supplyKeeper.SetModuleAccount(ctx, stakedPool)
	} else {
		if !stakedPool.GetCoins().IsEqual(stakedCoins) {
			panic(fmt.Sprintf("%s module account total does not equal the amount in each application account", types.StakedPoolName))
		}
	}
	// set the params set in the keeper
	keeper.Paramstore.SetParamSet(ctx, &data.Params)
}

// ExportGenesis returns a GenesisState for a given context and keeper
func ExportGenesis(ctx sdk.Context, keeper keeper.Keeper) types.GenesisState {
	params := keeper.GetParams(ctx)
	applications := keeper.GetAllApplications(ctx)
	return types.GenesisState{
		Params:       params,
		Applications: applications,
		Exported:     true,
	}
}

// ValidateGenesis validates the provided staking genesis state to ensure the
// expected invariants holds. (i.e. params in correct bounds, no duplicate applications)
func ValidateGenesis(data types.GenesisState) error {
	err := validateGenesisStateApplications(data.Applications, sdk.NewInt(data.Params.AppStakeMin))
	if err != nil {
		return err
	}
	err = data.Params.Validate()
	if err != nil {
		return err
	}
	return nil
}

func validateGenesisStateApplications(applications []types.Application, minimumStake sdk.Int) (err error) {
	addrMap := make(map[string]bool, len(applications))
	for i := 0; i < len(applications); i++ {
		app := applications[i]
		strKey := app.PublicKey.RawString()
		if _, ok := addrMap[strKey]; ok {
			return fmt.Errorf("duplicate application in genesis state: address %v", app.GetAddress())
		}
		if app.Jailed && app.IsStaked() {
			return fmt.Errorf("application is staked and jailed in genesis state: address %v", app.GetAddress())
		}
		if app.StakedTokens.IsZero() && !app.IsUnstaked() {
			return fmt.Errorf("staked/unstaked genesis application cannot have zero stake, application: %v", app)
		}
		addrMap[strKey] = true
		if !app.IsUnstaked() && app.StakedTokens.LTE(minimumStake) {
			return fmt.Errorf("application has less than minimum stake: %v", app)
		}
	}
	return
}
