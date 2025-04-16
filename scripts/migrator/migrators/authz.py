from migrators import Migrator


class Authz(Migrator):
    # Authz just makes the state remain untouched
    def migrate(self, data: dict):
        data["authz"] = {"authorization": []}
