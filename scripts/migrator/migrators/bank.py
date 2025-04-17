from migrators import Migrator
from migrators.utils.utils import coins_to_decimals

BAD_ADDRESS="kii1qqqqqta26mc9wa6zvzmv9jccv8ugunj7jmgqfl"
REWARDS="kii1ayzckayhqvr3ujd5qn58avy7t85mjye3gc40fh"
ONE_TOKEN=1_000_000_000_000_000_000

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

        # Clear a bad address on the system
        self.clear_bad_address(bank["balances"])

    # Migrate the params
    # Params are the same
    def migrate_params(self, params: dict):
        return

    # Migrate the denom metadata
    # Denom metadata remains the same
    def migrate_denom_metadata(self, denom_metadata: list[dict]):
        return

    # Migrate the supply
    # Update akii to have 12 more decimals
    def migrate_supply(self, supply: list[dict]):
        # Convert the akii on the supply to 18 decimals
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

            # Now iterate the coins and update only the akii value
            for coin in balance["coins"]:
                denom = coin["denom"]
                amount = coin["amount"]

                # Check if akii
                if denom == "akii":
                    coin["amount"] = f"{amount}{wei_balance}"
            
            # Sort the list
            balance["coins"].sort(key=lambda c: c["denom"])

        # Add new balances from wei_balances that are missing
        for address, amount in wei_balances_dict.items():
            if address not in address_to_balance:
                new_balance = {
                    "address": address,
                    "coins": [{"denom": "akii", "amount": amount}],
                }
                balance_data.append(new_balance)
        
    # Clear a bad balance on the chain holding tokens
    def clear_bad_address(self, balance_data: list[dict]):
        # Iterate all the addresses and find the bad address
        remainder=0
        for balance in balance_data:
            address = balance["address"]

            # Check if it's the bad address
            if address == BAD_ADDRESS:
                # Iterate the coins and reduce the Kii amount
                for coin in balance["coins"]:
                    denom = coin["denom"]
                    amount = coin["amount"]

                    # Check if akii
                    if denom == "akii":
                        amount_int = int(coin["amount"])

                        # Leave the address with one token
                        if amount_int > ONE_TOKEN:
                            coin["amount"] = f"{ONE_TOKEN}"
                            remainder = amount_int - ONE_TOKEN
                            print(f"Removed {remainder} from address {address}")
        
        # Iterate again and add to the rewards address
        if remainder == 0:
            return
        
        for balance in balance_data:
            address = balance["address"]

            # Check if it's the rewards address
            if address == REWARDS:
                # Iterate the coins and reduce the Kii amount
                for coin in balance["coins"]:
                    denom = coin["denom"]
                    amount = coin["amount"]

                    # Check if akii
                    if denom == "akii":
                        amount_int = int(coin["amount"])

                        coin["amount"] = f"{amount_int+remainder}"
                        print(f"Added {amount_int+remainder} from address {address}")
