from migrators import Migrator
from migrators.utils.utils import coins_to_decimals


class Slashing(Migrator):
    def migrate(self, data: dict):
        slashing = data["slashing"]

        # Migrate the params
        self.migrate_params(slashing["params"])
        # Migrate the signing infos
        self.migrate_signing_infos(slashing["signing_infos"])
        # Migrate the missed blocks
        self.migrate_missed_blocks(slashing["missed_blocks"])

        return

    # Migrate the params
    # We do no changes
    def migrate_params(self, params: dict):
        return

    # Migrate the signing infos
    # We do no changes
    def migrate_signing_infos(self, signing_infos: list[dict]):
        return

    # Migrate the missed blocks
    # Clear the missed blocks and remove window_size
    def migrate_missed_blocks(self, missed_blocks: list[dict]):
        for missed_block in missed_blocks:
            missed_block["missed_blocks"] = []
            del missed_block["window_size"]
