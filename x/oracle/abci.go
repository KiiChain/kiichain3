package oracle

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kiichain/kiichain3/x/oracle/keeper"
	"github.com/kiichain/kiichain3/x/oracle/types"
	"github.com/kiichain/kiichain3/x/oracle/utils"
)

// MidBlocker is the function executed when each block has been completed
// this function get the votes from the validators, calculate the exchange rate using
// weighted median logic when the vote period is almost finished
func MidBlocker(ctx sdk.Context, k keeper.Keeper) {
	params := k.GetParams(ctx)

	// Check if the current block is the last one to finish the voting period
	if utils.IsPeriodLastBlock(ctx, params.VotePeriod) {
		validatorClaimMap := make(map[string]types.Claim)

		maxValidators := k.StakingKeeper.MaxValidators(ctx) // get amount of bonded validators
		iterator := k.StakingKeeper.ValidatorsPowerStoreIterator(ctx)
		defer iterator.Close()

		powerReduction := k.StakingKeeper.PowerReduction(ctx) // Get the power reduction factor

		// Get the validators ordered by power
		powerOrderedValAddrs := []sdk.ValAddress{}
		for i := 0; iterator.Valid() && i < int(maxValidators); iterator.Next() {
			powerOrderedValAddrs = append(powerOrderedValAddrs, iterator.Value())
		}

		// Iterate over the validators and exlude not bonded
		for _, valAddr := range powerOrderedValAddrs {
			validator := k.StakingKeeper.Validator(ctx, valAddr) // Get the validator by its address

			// exclude not bonded validators
			if validator.IsBonded() {
				valAddr := validator.GetOperator()                      // Get operator address
				valPower := validator.GetConsensusPower(powerReduction) // Get validator's power
				claim := types.NewClaim(valPower, 0, 0, false, valAddr) // Create claim object
				validatorClaimMap[valAddr.String()] = claim             // Assign the validator on the list to receive
			}

		}

		// Get the voting targets from the KVStore
		voteTargets := make(map[string]types.Denom)
		totalTargets := 0
		k.IterateVoteTargets(ctx, func(denom string, denomInfo types.Denom) bool {
			voteTargets[denom] = denomInfo
			totalTargets++
			return false
		})

		// Create a reference denom (RD) based on the voting power
		voteMap := k.OrganizeBallotByDenom(ctx, validatorClaimMap) // Create a map (denom sorted) with the votes by denom
		referenceDenom, belowThresholdVoteMap := pickReferenceDenom(ctx, k, voteTargets, voteMap)

		if referenceDenom != "" {
			ballotRD := voteMap[referenceDenom] // get the ballot of the RD
			votingMapRD := ballotRD.ToMap()     // Conver the ballot into a map by voting tally

			// calculate the weighted median of the reference denom ballot
			exchangeRateRD := ballotRD.WeightedMedianWithAssertion()

			// Get the denoms from the ballot
			denoms := make([]string, len(voteMap))
			index := 0
			for denom := range voteMap {
				denoms[index] = denom
				index++
			}

			sort.Strings(denoms)

			// Iterate the denoms on the voting map to ....
			for _, denom := range denoms {
				votingTally := voteMap[denom] // get the voting tally per denom

				// Convert the voting tally to cross exchange rate
				if denom != referenceDenom {
					votingTally = votingTally.ToCrossRateWithSort(votingMapRD)
				}

				// Get weighted median of cross exchange rates
				exchangeRate := Tally(ctx, votingTally, params.RewardBand, validatorClaimMap)

				// Validate invalid exchangeRate
				if exchangeRate.IsZero() {
					continue // skip this denom
				}

				// transform into the original form base/quote
				if denom != referenceDenom {
					exchangeRate = exchangeRateRD.Quo(exchangeRate)
				}

				// set the exchange rate with event
				k.SetBaseExchangeRateWithEvent(ctx, denom, exchangeRate)
			}
		}

		// Extract the denoms stored on belowThresholdVote map
		belowThresholdDenoms := make([]string, len(belowThresholdVoteMap))
		n := 0
		for denom := range belowThresholdVoteMap {
			belowThresholdDenoms[n] = denom
			n++
		}
		sort.Strings(belowThresholdDenoms) // sort by denom name

		// Calculate tally for below threshold assets lists
		for _, denom := range belowThresholdDenoms {
			ballot := belowThresholdVoteMap[denom]
			Tally(ctx, ballot, params.RewardBand, validatorClaimMap)
		}

		// ********* Miss Counting Logic *************
		// TODO: Complete this part

	}
}
