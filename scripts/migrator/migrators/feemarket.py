from migrators import Migrator


class FeeMarket(Migrator):
    # Migrate just set the default state
    def migrate(self, data: dict):
        data["feemarket"] = {
            "block_gas": "96479",
            "params": {
                "base_fee": "392695903.778098200015595140",
                "base_fee_change_denominator": 8,
                "elasticity_multiplier": 2,
                "enable_height": "0",
                "min_gas_multiplier": "0.500000000000000000",
                "min_gas_price": "0.000000000000000000",
                "no_base_fee": False,
            },
        }
