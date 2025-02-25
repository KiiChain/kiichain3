package keeper

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kiichain/kiichain3/x/oracle/types"
)

// OrganizeBallotByDenom iterates over the map with validators and create its voting tally.
// returns a map with the denom and its ballot (denom alphabetical ordered)
func (k Keeper) OrganizeBallotByDenom(ctx sdk.Context, validatorClaimMap map[string]types.Claim) map[string]types.ExchangeRateBallot {
	votes := map[string]types.ExchangeRateBallot{} // Here I will collect the array of votes by denom

	// Aggregate votes by denom
	aggregateHandler := func(voterAddr sdk.ValAddress, aggregateVote types.AggregateExchangeRateVote) bool {
		// Aggregate only for validators who have registered on the map
		claim, ok := validatorClaimMap[aggregateVote.Voter]

		if ok {
			power := claim.Power
			for _, tuple := range aggregateVote.ExchangeRateTuples {
				tmpPower := power

				// Validate invalids exchange rates
				if !tuple.ExchangeRate.IsPositive() {
					tmpPower = 0
				}

				vote := types.NewVoteForTally(tuple.ExchangeRate, tuple.Denom, voterAddr, tmpPower) // Create validator vote
				votes[tuple.Denom] = append(votes[tuple.Denom], vote)                               // Append vote on that specific denom
			}
		}
		return false
	}

	k.IterateAggregateExchangeRateVotes(ctx, aggregateHandler)

	// sort created ballot
	for denom, ballot := range votes {
		sort.Sort(ballot) // sort by denom
		votes[denom] = ballot
	}

	return votes
}
