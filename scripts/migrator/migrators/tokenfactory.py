from migrators import Migrator


class TokenFactory(Migrator):
    def migrate(self, data: dict):
        tokenfactory = data["tokenfactory"]

        # Migrate the params
        tokenfactory["params"] = {
            "denom_creation_fee": [{"denom": "akii", "amount": "10000000"}],
            "denom_creation_gas_consume": "2000000",
        }

        return
