from migrators import Migrator
from migrators.utils.utils import coins_to_decimals


class Gov(Migrator):
    def migrate(self, data: dict):
        gov = data["gov"]

        # Migrate the params
        gov["params"] = self.migrate_params(gov)

        # Clear the old proposals
        gov["starting_proposal_id"] = "1"
        gov["deposits"] = []
        gov["votes"] = []
        gov["proposals"] = []

        # Clear the old params
        gov["deposit_params"] = None
        gov["voting_params"] = None
        gov["tally_params"] = None

        # Add the new param
        gov["constitution"] = ""

        return

    # Migrate the params
    def migrate_params(self, gov: dict) -> dict:
        # Get the params from the original module
        deposit_params = gov["deposit_params"]
        tally_params = gov["tally_params"]
        voting_params = gov["voting_params"]

        # Get the new min_deposit
        min_deposit = deposit_params["min_deposit"]
        coins_to_decimals(min_deposit)

        # Get the expedited deposit
        min_expedited_deposit = deposit_params["min_expedited_deposit"]
        coins_to_decimals(min_expedited_deposit)

        return {
            "min_deposit": min_deposit,
            "max_deposit_period": deposit_params["max_deposit_period"],
            "voting_period": voting_params["voting_period"],
            "quorum": tally_params["quorum"],
            "threshold": tally_params["threshold"],
            "veto_threshold": tally_params["veto_threshold"],
            "min_initial_deposit_ratio": "0.000000000000000000",
            "proposal_cancel_ratio": "0.500000000000000000",
            "proposal_cancel_dest": "",
            "expedited_voting_period": voting_params["expedited_voting_period"],
            "expedited_threshold": tally_params["expedited_threshold"],
            "expedited_min_deposit": min_expedited_deposit,
            "burn_vote_quorum": False,
            "burn_proposal_deposit_prevote": False,
            "burn_vote_veto": True,
            "min_deposit_ratio": "0.010000000000000000",
        }
