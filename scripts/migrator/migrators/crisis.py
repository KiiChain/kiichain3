from migrators import Migrator
from migrators.utils.utils import coins_to_decimals, add_decimals


class Crisis(Migrator):
    # Migrate just set the default state
    def migrate(self, data: dict):
        crisis = data["crisis"]

        # Migrate the constant_fee
        constant_fee = crisis["constant_fee"]
        constant_fee['amount'] = add_decimals(constant_fee['amount'])
