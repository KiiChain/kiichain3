from migrators import Migrator
from migrators.utils.utils import dec_coins_to_decimals, dec_add_decimals


class Distribution(Migrator):
    # Migrate the accounts and the params
    def migrate(self, data: dict):
        distribution = data["distribution"]

        # Migrate the params
        self.migrate_params(distribution["params"])
        # Migrate the fee pool
        self.migrate_fee_pool(distribution["fee_pool"])
        # Migrate the withdraw infos
        self.migrate_delegator_withdraw_infos(distribution["delegator_withdraw_infos"])
        # Migrate the previous proposer
        self.migrate_previous_proposer(distribution["previous_proposer"])
        # Migrate the outstanding_reward
        self.migrate_outstanding_rewards(distribution["outstanding_rewards"])
        # Migrate the validator accumulated commissions
        self.migrate_validator_accumulated_commissions(
            distribution["validator_accumulated_commissions"]
        )
        # Migrate the validator historical rewards
        self.migrate_validator_historical_rewards(
            distribution["validator_historical_rewards"]
        )
        # Migrate the validator current rewards
        self.migrate_validator_current_rewards(
            distribution["validator_current_rewards"]
        )
        # Migrate the delegator starting infos
        self.migrate_delegator_starting_infos(distribution["delegator_starting_infos"])
        # Migrate the validator slash events
        self.migrate_validator_slash_events(distribution["validator_slash_events"])

    # Migrate the params
    # Do nothing
    def migrate_params(self, data: dict):
        return

    # Migrate the fee pool
    # Get the pool to 18 decimals
    def migrate_fee_pool(self, data: dict):
        community_pool: list[dict] = data["community_pool"]

        # Migrate the community pool
        dec_coins_to_decimals(community_pool)

    # Migrate the delegators withdraw infos
    def migrate_delegator_withdraw_infos(self, data: list):
        return

    # Migrate the previous proposer
    # Do nothing
    def migrate_previous_proposer(self, data: str):
        return

    # Migrate the outstanding reward
    def migrate_outstanding_rewards(self, outstanding_rewards: list[dict]):
        # Iterate all the entries
        for outstanding_reward in outstanding_rewards:
            # Update the outstanding reward
            validator_outstanding_rewards = outstanding_reward["outstanding_rewards"]
            # Convert to 18 decimals
            dec_coins_to_decimals(validator_outstanding_rewards)

    # Migrate the validator accumulated commissions
    def migrate_validator_accumulated_commissions(
        self, validator_accumulated_commissions: list[dict]
    ):
        # Iterate all the accumulated commissions
        for accumulated_commission in validator_accumulated_commissions:
            # Get the accumulated commission dec coins
            accumulated_commission_dec_coins = accumulated_commission["accumulated"][
                "commission"
            ]

            # Update the commission to 18 decimals
            dec_coins_to_decimals(accumulated_commission_dec_coins)

    # Migrate the validator historical rewards
    # This one is fine, we don't migrate
    def migrate_validator_historical_rewards(
        self, validator_historical_rewards: list[dict]
    ):
        return

    # Migrate the validator current rewards
    def migrate_validator_current_rewards(self, validator_current_rewards: list[dict]):
        # Iterate all the current rewards
        for validator_current_reward in validator_current_rewards:
            # Get the reward
            rewards = validator_current_reward["rewards"]["rewards"]
            # Update the decimals there
            dec_coins_to_decimals(rewards)

    # Migrate the validator starting infos
    def migrate_delegator_starting_infos(self, delegator_starting_infos: list[dict]):
        # Iterate all the staking infos
        for delegator_starting_info in delegator_starting_infos:
            # Get the starting info
            starting_info = delegator_starting_info["starting_info"]

            # Update the stake
            starting_info["stake"] = dec_add_decimals(starting_info["stake"])

    # Migrate the validator slash events
    def migrate_validator_slash_events(self, validator_slash_events: list[dict]):
        return
