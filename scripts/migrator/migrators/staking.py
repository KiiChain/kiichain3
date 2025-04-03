from migrators import Migrator
from migrators.utils.utils import dec_add_decimals, add_decimals


class Staking(Migrator):
    def migrate(self, data: dict):
        staking = data["staking"]

        # Migrate the params
        self.migrate_params(staking["params"])
        # Migrate the last total power
        self.migrate_last_total_power(staking["last_total_power"])
        # Migrate the last validator_powers
        self.migrate_last_validator_powers(staking['last_validator_powers'])
        # Migrate the validators
        self.migrate_validators(staking['validators'])
        # Migrate the delegations
        self.migrate_delegations(staking['delegations'])
        # Migrate the unbonding delegations
        self.migrate_unbonding_delegations(staking['unbonding_delegations'])
        # Migrate the redelegations
        self.migrate_redelegations(staking['redelegations'])
        # Migrate the exported
        self.migrate_exported(staking['exported'])

        print("CHECK POWER CALCULATION, MAYBE SHOULD ADD THE 12 DECIMALS")

        return

    # Migrate the params
    def migrate_params(self, params: dict):
        # delete the unused params
        del(params["max_voting_power_enforcement_threshold"])
        del(params["max_voting_power_ratio"])

    # Migrate the last total power
    # We do no change
    def migrate_last_total_power(self, last_total_power: str):
        return
    
    # Migrate the last validator powers
    # We do no change
    def migrate_last_validator_powers(self, last_validator_powers: list[dict]):
        return
    
    # Migrate the validators
    def migrate_validators(self, validators: list[dict]):
        # Iterate all the validators
        print("CHECK unbonding_on_hold_ref_count")
        for validator in validators:
            validator['unbonding_on_hold_ref_count'] = "0"
            validator['unbonding_ids'] = []

            # Update the decimals
            validator['delegator_shares'] = dec_add_decimals(validator['delegator_shares'])
            validator['tokens'] = add_decimals(validator['tokens'])


    # Migrate the delegations
    def migrate_delegations(self, delegations: list[dict]):
        # Iterate all the delegations
        for delegation in delegations:
            delegation['shares'] = dec_add_decimals(delegation['shares'])
    
    # Migrate the unbonding delegations
    def migrate_unbonding_delegations(self, unbonding_delegations: list[dict]):
        return

    # Migrate the redelegations
    def migrate_redelegations(self, redelegations: list[dict]):
        return
    
    # Migrate the exported
    def migrate_exported(self, exported: bool):
        return