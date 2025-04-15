from migrators import Migrator


class FeeIBC(Migrator):
    # Migrate just set the default state
    def migrate(self, data: dict):
        data["feeibc"] = {
            "identified_fees": [],
            "fee_enabled_channels": [],
            "registered_payees": [],
            "registered_counterparty_payees": [],
            "forward_relayers": []
        }
