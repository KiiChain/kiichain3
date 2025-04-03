from migrators import Migrator

class Evidence(Migrator):
    # Evidence just makes the state remain untouched
    def migrate(self, data: dict):
        capability = data['evidence']

        print("CHECK ME EVIDENCE MIGRATOR")

        # We do no changes to evidence
        return

