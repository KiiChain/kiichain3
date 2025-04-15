from migrators import Migrator


class PacketForwardMiddleware(Migrator):
    # Migrate just set the default state
    def migrate(self, data: dict):
        data["packetfowardmiddleware"] = {"in_flight_packets": {}}
