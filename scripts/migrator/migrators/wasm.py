from migrators import Migrator
from migrators.utils.utils import base64_to_hex


class Wasm(Migrator):
    def migrate(self, data: dict):
        wasm = data["wasm"]

        # Migrate the codes
        self.migrate_codes(wasm["codes"])

        # Migrate the contracts
        self.migrate_contracts(wasm["contracts"])

        # Migrate the params
        self.migrate_params(wasm["params"])

        # Delete the gen msgs
        del wasm["gen_msgs"]

        return

    # Migrate params and change the addresses to a list
    def migrate_params(self, params: dict):
        code_upload_access = params["code_upload_access"]

        # Change addresses to a empty list
        code_upload_access["addresses"] = []

        # Delete address
        del code_upload_access["address"]

    # Migrate the codes
    def migrate_codes(self, codes: list[dict]):
        # Iterate all the codes
        for code in codes:
            # Get the instantiate config
            instantiate_config = code["code_info"]["instantiate_config"]

            # Change addresses to a empty list
            instantiate_config["addresses"] = []

            # Delete address
            del instantiate_config["address"]

    # Migrate the contracts
    def migrate_contracts(self, contracts: list[dict]):
        # Iterate all the contracts
        for contract in contracts:
            states = contract["contract_state"]

            # Get the contract info
            contract_info = contract["contract_info"]

            # Add the created
            contract_info["created"] = {"block_height": "0", "tx_index": "0"}

            # Add the code history
            contract["contract_code_history"] = [
                {
                    "operation": "CONTRACT_CODE_HISTORY_OPERATION_TYPE_INIT",
                    "code_id": contract_info["code_id"],
                    "updated": {"block_height": "0", "tx_index": "0"},
                    "msg": {"count": 1},
                }
            ]

            # Iterate all the contract states
            for state in states:
                # Update the keys with hex
                state["key"] = base64_to_hex(state["key"])
