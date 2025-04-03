from migrators import Migrator

class Deleter(Migrator):
    def migrate(self, data: dict):
        # Delete itself
        return

    def delete_self(self):
        return True