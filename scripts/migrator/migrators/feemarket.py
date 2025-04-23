from migrators import Migrator


class FeeMarket(Migrator):
    # Migrate just set the default state
    def migrate(self, data: dict):
        data["feemarket"] = {
            "params": {
                "no_base_fee": False,
                "base_fee_change_denominator": 8,
                "elasticity_multiplier": 2,
                "enable_height": "0",
                "base_fee": "1000000000.000000000000000000",
                "min_gas_price": "0.000000000000000000",
                "min_gas_multiplier": "0.500000000000000000",
            },
            "block_gas": "0",
        }
