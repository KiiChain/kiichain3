from collections import defaultdict

from migrators import Migrator
from migrators.auth import Auth
from migrators.authz import Authz
from migrators.bank import Bank
from migrators.distribution import Distribution
from migrators.evidence import Evidence
from migrators.evm import EVM
from migrators.feegrant import Feegrant
from migrators.gov import Gov
from migrators.ibc import IBC
from migrators.slashing import Slashing
from migrators.staking import Staking
from migrators.transfer import Transfer
from migrators.crisis import Crisis
from migrators.ratelimit import RateLimit
from migrators.wasm import Wasm
from migrators.tokenfactory import TokenFactory

from migrators.erc20 import ERC20
from migrators.feemarket import FeeMarket
from migrators.feeIBC import FeeIBC
from migrators.interchainaccounts import InterchainAccounts
from migrators.packetfowardmiddleware import PacketForwardMiddleware

from migrators.delete import Deleter
from migrators.nochange import NoChange
from migrators.utils.address_converter import convert_bech32_prefix, hex_to_bech32
from migrators.utils.utils import replace_in_dict

# Define all the migrators
MIGRATORS: dict[str, Migrator] = {
    "accesscontrol": Deleter(),
    "auth": Auth(),
    "authz": Authz(),
    "bank": Bank(),
    "capability": NoChange(),
    "crisis": Crisis(),
    "distribution": Distribution(),
    "epoch": Deleter(),
    "evidence": Evidence(),
    "evm": EVM("akii", "3665", "KII", evm_decimals="18"),
    "feegrant": Feegrant(),
    "genutil": NoChange(),
    "gov": Gov(),
    "ibc": IBC(),
    "mint": Deleter(),
    "params": Deleter(),
    "slashing": Slashing(),
    "staking": Staking(),
    "tokenfactory": TokenFactory(),
    "transfer": Transfer(),
    "upgrade": NoChange(),
    "vesting": NoChange(),
    "wasm": Wasm(),
}


# Define all modules to add
MODULES_TO_ADD: list[Migrator] = {
    ERC20(),
    FeeMarket(),
    FeeIBC(),
    InterchainAccounts(),
    PacketForwardMiddleware(),
    RateLimit(),
}


# Root migrator migrates the root of the application
# This migrates the following fields:
# - app_name (as a static variable)
# - app_version (from consensus_params)
# - genesis_time
# - chain_id
# - initial_height (From string to int)
# - app_hash (From empty string to null)
# - consensus (From consensus_params)
# - app_state
class Root(Migrator):
    def migrate(self, data: dict):
        # Migrate the app_name
        data["app_name"] = "kiichaind"
        # Migrate the app_version
        data["app_version"] = "1.0.0"  # TODO: CHECK ME
        # Migrate the genesis_time
        data["genesis_time"] = data["genesis_time"]
        # Migrate the chain_id
        data["chain_id"] = "kiichain_1336-1"
        # Migrate the initial_height
        data["initial_height"] = int(data["initial_height"])
        # Migrate the app_hash
        data["app_hash"] = None

        # Migrate the addresses
        data = self.migrate_addresses(data)

        # Migrate the denom
        data = self.migrate_denom(data)

        # Migrate the consensus
        data["consensus"] = self.migrate_consensus(data)
        del data["consensus_params"]
        del data["validators"]

        # Migrate the app_state
        self.migrate_app_state(data["app_state"])

        return data

    # Migrates the app state
    # Iterate each of the keys and run it's own migration function
    def migrate_app_state(self, data: dict):
        # Iterate all the modules
        for key in list(data.keys()):
            # Delete not needed keys
            if key not in MIGRATORS.keys():
                raise Exception(f"Module {key} migration function not found")

            # Delete if needed
            if MIGRATORS[key].delete_self():
                del data[key]
                continue

            # Run the migration
            MIGRATORS[key].migrate(data)

            # Print the class name
            MIGRATORS[key].print_class_name()

        # Iterate all the modules to add
        for module in MODULES_TO_ADD:
            module.migrate(data)

    # Migrate the consensus param from the chain
    # This migrates the following fields:
    # - validators (from the root of the genesis)
    # - block (Similar to https://cosmoshub.lava.build/cosmos/consensus/v1/params)
    # - evidence (As the default params)
    # - validator
    # - version (param app_version changed to app)
    # - abci (only the param vote_extensions_enable_height)
    def migrate_consensus(self, data: dict) -> dict:
        # Start a new data for the response
        consensus = {}

        # Separate the consensus params for easy access
        consensus_params = data["consensus_params"]

        # Migrate the validators
        consensus["validators"] = data["validators"]

        # Migrate the params
        consensus["params"] = {}
        consensus["params"]["block"] = consensus_params["block"]
        consensus["params"]["evidence"] = {
            "max_age_num_blocks": "100000",
            "max_age_duration": "172800000000000",
            "max_bytes": "1048576",
        }
        consensus["params"]["validator"] = consensus_params["validator"]
        consensus["params"]["version"] = {
            "app": consensus_params["version"]["app_version"]
        }
        consensus["params"]["abci"] = {
            "vote_extensions_enable_height": str(
                consensus_params["abci"]["vote_extensions_enable_height"]
            )
        }

        return consensus

    # Migrate the addresses
    def migrate_addresses(self, data) -> dict:
        # Convert all the contract addresses
        evm_data = data["app_state"]["evm"]

        # Generate a list of the validator addresses
        non_migrate_addresses = defaultdict(bool)
        for validator in data["app_state"]["staking"]["validators"]:
            operator_address = validator["operator_address"]
            converted_operator_address = convert_bech32_prefix(operator_address, "kii")

            # Add to the list of addresses to not migrate
            non_migrate_addresses[converted_operator_address] = "validator"

        # Check all the module accounts
        # We don't migrate module accounts
        for account in data["app_state"]["auth"]["accounts"]:
            account_type = account["@type"]

            if account_type == "/cosmos.auth.v1beta1.ModuleAccount":
                address = account["base_account"]["address"]
                non_migrate_addresses[address] = "module account"

        # Check all the wasm addresses
        for wasm_contract in data["app_state"]["wasm"]["contracts"]:
            contract_address = wasm_contract["contract_address"]
            non_migrate_addresses[contract_address] = "wasm contract"

        # Lets not migrate any of the genesis addresses (any account bellow account number 50)
        for account in data["app_state"]["auth"]["accounts"]:
            acc_number = 0
            address = ""
            # Check if it has a base account
            if "base_account" in account.keys():
                address = account["base_account"]["address"]
                acc_number = int(account["base_account"]["account_number"])
            else:
                address = account["address"]
                acc_number = int(account["account_number"])

            # Check if we should not migrate
            if acc_number and acc_number != 0 and acc_number <= 50:
                non_migrate_addresses[address] = f"account number ({acc_number})"

        # Iterate all the address associations
        replace_map = {}
        for association in evm_data["address_associations"]:
            kii_address = association["kii_address"]
            eth_address = association["eth_address"]

            # Check if it's a validator we don't convert
            if non_migrate_addresses[kii_address]:
                reason = non_migrate_addresses[kii_address]
                print(f"Address {kii_address} will not be migrated - reason: {reason}")
                continue

            # Convert the address
            processed_kii_address = hex_to_bech32(eth_address, "kii")

            # Add to the replace map
            replace_map[kii_address] = processed_kii_address

        # Apply
        data = replace_in_dict(data, replace_map)

        return data

    # Migrate the denom
    def migrate_denom(self, data) -> dict:
        # Apply
        replace_map = {"ukii": "akii"}
        data = replace_in_dict(data, replace_map)
        return data
