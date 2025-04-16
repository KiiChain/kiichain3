from migrators import Migrator


class Feegrant(Migrator):
    def migrate(self, data: dict):
        data["feegrant"] = {"allowances": []}
