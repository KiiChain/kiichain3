from migrators import Migrator


class ERC20(Migrator):
    # Migrate just set the default state
    def migrate(self, data: dict):
        data["erc20"] = {
            "params": {
                "dynamic_precompiles": [],
                "enable_erc20": True,
                "native_precompiles": [],
            },
            "token_pairs": [],
        }
