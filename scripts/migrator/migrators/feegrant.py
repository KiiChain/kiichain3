from migrators import Migrator

class Feegrant(Migrator):
    def migrate(self, data: dict):
        feegrant = data['feegrant']

        print("CHECK ME FEEGRANT MIGRATOR")

        return

