from migrators import Migrator


class Mint(Migrator):
    # Set the param as default
    def migrate(self, data: dict):
        data["mint"] = {
            "minter": {
                "inflation": "0.130000576714590760",
                "annual_provisions": "13000584903742401444318282.319415012646067840",
            },
            "params": {
                "mint_denom": "ukii",
                "inflation_rate_change": "0.130000000000000000",
                "inflation_max": "0.200000000000000000",
                "inflation_min": "0.070000000000000000",
                "goal_bonded": "0.670000000000000000",
                "blocks_per_year": "6311520",
            },
        }
