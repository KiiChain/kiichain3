from migrators import Migrator

class NoChange(Migrator):
    # NoChange applies no change to a module
    def migrate(self, data: dict):
        return

