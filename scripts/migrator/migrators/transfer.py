from migrators import Migrator


class Transfer(Migrator):
    def migrate(self, data: dict):
        transfer = data["transfer"]

        # Add the total escrowed param
        transfer['total_escrowed'] = []

        return
