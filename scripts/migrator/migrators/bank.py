from migrators import Migrator
from migrators.utils.utils import coins_to_decimals


class Bank(Migrator):
    # Migrate the accounts and the params
    def migrate(self, data: dict):
        bank = data["bank"]

        # Migrate the balances
        self.migrate_balances(bank["balances"], bank["wei_balances"])
        # Migrate the params
        self.migrate_params(bank["params"])
        # Migrate the denom metadata
        self.migrate_denom_metadata(bank["denom_metadata"])
        # Migrate the supply
        self.migrate_supply(bank["supply"])

        # Add the new keys
        bank["send_enabled"] = []

        # Delete the unused key
        del bank["wei_balances"]

    # Migrate the params
    # Params are the same
    def migrate_params(self, params: dict):
        return

    # Migrate the denom metadata
    # Denom metadata remains the same
    def migrate_denom_metadata(self, denom_metadata: list[dict]):
        return

    # Migrate the supply
    # Update ukii to have 12 more decimals
    def migrate_supply(self, supply: list[dict]):
        # Convert the ukii on the supply to 18 decimals
        coins_to_decimals(supply)

    # Put all balances to 18 decimals and delete the WEI balance key
    def migrate_balances(self, balance_data: list[dict], wei_balances: list[dict]):
        # We build a hash map of the Wei balances
        wei_balances_dict = {wb["address"]: wb["amount"] for wb in wei_balances}
        address_to_balance = {b["address"]: b for b in balance_data}

        # Now iterate all the balances
        for balance in balance_data:
            address = balance["address"]

            # Find if we have a wei balance to that address
            wei_balance = wei_balances_dict.get(address)

            if wei_balance is None:
                wei_balance = ""

            # Zero fill to 12 digits
            wei_balance = wei_balance.zfill(12)

            # Now iterate the coins and update only the ukii value
            for coin in balance["coins"]:
                denom = coin["denom"]
                amount = coin["amount"]

                # Check if ukii
                if denom == "ukii":
                    coin["amount"] = f"{amount}{wei_balance}"

        # Add new balances from wei_balances that are missing
        for address, amount in wei_balances_dict.items():
            if address not in address_to_balance:
                new_balance = {
                    "address": address,
                    "coins": [{"denom": "ukii", "amount": amount}],
                }
                balance_data.append(new_balance)
