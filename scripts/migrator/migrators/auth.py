from migrators import Migrator


class Auth(Migrator):
    # Migrate the accounts and the params
    def migrate(self, data: dict):
        auth = data["auth"]

        self.migrate_accounts(auth["accounts"])
        self.migrate_params(auth["params"])

    # Migrate the accounts
    def migrate_accounts(self, accounts: list[dict]):
        # Clear out pubkeys
        for account in accounts:
            # Check if it has pub_key as the root keys
            if "pub_key" not in account.keys():
                continue

            # Clear the pubkey
            account["pub_key"] = None

    # Migrate the params
    # Delete the key disable_seqno_check
    def migrate_params(self, data: dict):
        del data["disable_seqno_check"]
