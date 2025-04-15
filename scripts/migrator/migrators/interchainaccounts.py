from migrators import Migrator


class InterchainAccounts(Migrator):
    # Migrate just set the default state
    def migrate(self, data: dict):
        data["interchainaccounts"] = {
            "controller_genesis_state": {
                "active_channels": [],
                "interchain_accounts": [],
                "ports": [],
                "params": {"controller_enabled": True},
            },
            "host_genesis_state": {
                "active_channels": [],
                "interchain_accounts": [],
                "port": "icahost",
                "params": {"host_enabled": True, "allow_messages": ["*"]},
            },
        }
