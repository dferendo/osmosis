package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

func (suite *KeeperTestSuite) TestSuperfluidAfterEpochEnd() {
	testCases := []struct {
		name             string
		validatorStats   []stakingtypes.BondStatus
		superDelegations []superfluidDelegation
		expRewards       sdk.Coins
	}{
		{
			"happy path with single validator and delegator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			sdk.Coins{},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			valAddrs := suite.SetupValidators(tc.validatorStats)

			denoms, poolIds := suite.SetupGammPoolsAndSuperfluidAssets([]sdk.Dec{sdk.NewDec(20), sdk.NewDec(20)})

			// Generate delegator addresses
			delAddrs := CreateRandomAccounts(1)
			intermediaryAccs, _ := suite.SetupSuperfluidDelegations(delAddrs, valAddrs, tc.superDelegations, denoms)
			suite.checkIntermediaryAccountDelegations(intermediaryAccs)

			// gamm swap operation before refresh
			acc1 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())
			pool, err := suite.App.GAMMKeeper.GetPool(suite.Ctx, poolIds[0])
			suite.Require().NoError(err)
			poolAssets := pool.GetAllPoolAssets()

			coins := sdk.Coins{sdk.NewInt64Coin(poolAssets[1].Token.Denom, 100000000000000)}
			err = simapp.FundAccount(suite.App.BankKeeper, suite.Ctx, acc1, coins)
			suite.Require().NoError(err)
			_, _, err = suite.App.GAMMKeeper.SwapExactAmountOut(
				suite.Ctx, acc1,
				poolIds[0], poolAssets[1].Token.Denom, poolAssets[1].Token.Amount,
				sdk.NewCoin(poolAssets[0].Token.Denom, poolAssets[0].Token.Amount.Quo(sdk.NewInt(4))))
			suite.Require().NoError(err)

			// run epoch actions
			suite.BeginNewBlock(true)

			// check lptoken twap value set
			newEpochTwap := suite.App.SuperfluidKeeper.GetOsmoEquivalentMultiplier(suite.Ctx, denoms[0])
			suite.Require().Equal(newEpochTwap.String(), "15.000000000000000000")

			// check delegation changes
			for _, acc := range intermediaryAccs {
				valAddr, err := sdk.ValAddressFromBech32(acc.ValAddr)
				suite.Require().NoError(err)
				delegation, found := suite.App.StakingKeeper.GetDelegation(suite.Ctx, acc.GetAccAddress(), valAddr)
				suite.Require().True(found)
				suite.Require().Equal(sdk.NewDec(7500000), delegation.Shares)
				// TODO: Check reward distribution
				// suite.Require().NotEqual(sdk.Coins{}, )
			}
		})
	}
}

// func (suite *KeeperTestSuite) TestOnStartUnlock() {
// 	testCases := []struct {
// 		name             string
// 		validatorStats   []stakingtypes.BondStatus
// 		superDelegations []superfluidDelegation
// 		unbondingLockIds []uint64
// 		expUnbondingErr  []bool
// 	}{
// 		{
// 			"with single validator and single superfluid delegation and single lockup unlock",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded},
// 			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}},
// 			[]uint64{1},
// 			[]bool{false},
// 		},
// 		{
// 			"with single validator and multiple superfluid delegations and single undelegation",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded},
// 			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}, {0, 0, "gamm/pool/1", 1000000}},
// 			[]uint64{1},
// 			[]bool{false},
// 		},
// 		{
// 			"with single validator and multiple superfluid delegations and multiple undelegation",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded},
// 			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}, {0, 0, "gamm/pool/1", 1000000}},
// 			[]uint64{1, 2},
// 			[]bool{false, false},
// 		},
// 		{
// 			"with multiple validators and multiple superfluid delegations and multiple undelegations",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
// 			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}, {0, 1, "gamm/pool/1", 1000000}},
// 			[]uint64{1, 2},
// 			[]bool{false, false},
// 		},
// 		{
// 			"undelegating not available lock id",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded},
// 			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}},
// 			[]uint64{2},
// 			[]bool{true},
// 		},
// 		{
// 			"try undelegating twice for same lock id",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded},
// 			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}},
// 			[]uint64{1, 1},
// 			[]bool{false, true},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc
// 		suite.Run(tc.name, func() {
// 			suite.SetupTest()

// 			poolId := suite.createGammPool([]string{appparams.BaseCoinUnit, "foo"})
// 			suite.Require().Equal(poolId, uint64(1))

// 			// Generate delegator addresses
// 			delAddrs := CreateRandomAccounts(1)

// 			// setup validators
// 			valAddrs := suite.SetupValidators(tc.validatorStats)

// 			// setup superfluid delegations
// 			intermediaryAccs, _ := suite.SetupSuperfluidDelegations(delAddrs, valAddrs, tc.superDelegations)
// 			suite.checkIntermediaryAccountDelegations(intermediaryAccs)

// 			for index, lockId := range tc.unbondingLockIds {
// 				// get intermediary account
// 				accAddr := suite.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.Ctx, lockId)
// 				intermediaryAcc := suite.App.SuperfluidKeeper.GetIntermediaryAccount(suite.Ctx, accAddr)
// 				valAddr := intermediaryAcc.ValAddr

// 				// unlock native lockup
// 				lock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, lockId)
// 				if err == nil {
// 					err = suite.App.LockupKeeper.BeginUnlock(suite.Ctx, *lock, nil)
// 				}

// 				if tc.expUnbondingErr[index] {
// 					suite.Require().Error(err)
// 					continue
// 				}
// 				suite.Require().NoError(err)

// 				// check lockId and intermediary account connection deletion
// 				addr := suite.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.Ctx, lockId)
// 				suite.Require().Equal(addr.String(), "")

// 				// check bonding synthetic lockup deletion
// 				_, err = suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, lockId, keeper.StakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
// 				suite.Require().Error(err)

// 				// check unbonding synthetic lockup creation
// 				unbondingDuration := suite.App.StakingKeeper.GetParams(suite.Ctx).UnbondingTime
// 				synthLock, err := suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, lockId, keeper.UnstakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
// 				suite.Require().NoError(err)
// 				suite.Require().Equal(synthLock.UnderlyingLockId, lockId)
// 				suite.Require().Equal(synthLock.SynthDenom, keeper.UnstakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
// 				suite.Require().Equal(synthLock.EndTime, suite.Ctx.BlockTime().Add(unbondingDuration))
// 			}
// 		})
// 	}
// }

func (suite *KeeperTestSuite) TestBeforeSlashingUnbondingDelegationHook() {
	testCases := []struct {
		name                  string
		validatorStats        []stakingtypes.BondStatus
		delegatorNumber       int
		superDelegations      []superfluidDelegation
		superUnbondingLockIds []uint64
		slashedValIndexes     []int64
		expSlashedLockIds     []uint64
		expUnslashedLockIds   []uint64
	}{
		{
			"happy path with single validator and multiple superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			1,
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]uint64{1},
			[]int64{0},
			[]uint64{1},
			[]uint64{},
		},
		{
			"with single validator and multiple superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			2,
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 0, 0, 1000000}},
			[]uint64{1, 2},
			[]int64{0},
			[]uint64{1, 2},
			[]uint64{},
		},
		{
			"with multiple validators and multiple superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			2,
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
			[]uint64{1, 2},
			[]int64{0},
			[]uint64{1},
			[]uint64{2},
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()

			slashFactor := sdk.NewDecWithPrec(5, 2)

			// Generate delegator addresses
			delAddrs := CreateRandomAccounts(tc.delegatorNumber)

			// setup validators
			valAddrs := suite.SetupValidators(tc.validatorStats)

			denoms, _ := suite.SetupGammPoolsAndSuperfluidAssets([]sdk.Dec{sdk.NewDec(20), sdk.NewDec(20)})

			// setup superfluid delegations
			intermediaryAccs, _ := suite.SetupSuperfluidDelegations(delAddrs, valAddrs, tc.superDelegations, denoms)
			suite.checkIntermediaryAccountDelegations(intermediaryAccs)

			for _, lockId := range tc.superUnbondingLockIds {
				lock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, lockId)
				suite.Require().NoError(err)

				// superfluid undelegate
				err = suite.App.SuperfluidKeeper.SuperfluidUndelegate(suite.Ctx, lock.Owner, lockId)
				suite.Require().NoError(err)
			}

			// slash unbonding lockups for all intermediary accounts
			for _, valIndex := range tc.slashedValIndexes {
				validator, found := suite.App.StakingKeeper.GetValidator(suite.Ctx, valAddrs[valIndex])
				suite.Require().True(found)
				suite.Ctx = suite.Ctx.WithBlockHeight(100)
				consAddr, err := validator.GetConsAddr()
				suite.Require().NoError(err)
				// slash by slash factor
				power := sdk.TokensToConsensusPower(validator.Tokens, sdk.DefaultPowerReduction)
				suite.App.StakingKeeper.Slash(suite.Ctx, consAddr, 80, power, slashFactor)
				// Note: this calls BeforeSlashingUnbondingDelegation hook
			}

			// check slashed lockups
			for _, lockId := range tc.expSlashedLockIds {
				gotLock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, lockId)
				suite.Require().NoError(err)
				suite.Require().Equal(sdk.NewInt(950000).String(), gotLock.Coins.AmountOf(denoms[0]).String())
			}

			// check unslashed lockups
			for _, lockId := range tc.expUnslashedLockIds {
				gotLock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, lockId)
				suite.Require().NoError(err)
				suite.Require().Equal(sdk.NewInt(1000000).String(), gotLock.Coins.AmountOf(denoms[0]).String())
			}
		})
	}
}
