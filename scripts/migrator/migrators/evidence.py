from migrators import Migrator


class Evidence(Migrator):
    # Evidence just makes the state remain untouched
    def migrate(self, data: dict):
        evidence = data["evidence"]

        # We do no changes to evidence
        return
