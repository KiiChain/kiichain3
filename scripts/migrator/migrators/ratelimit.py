from migrators import Migrator


class RateLimit(Migrator):
    # Migrate just set the default state
    def migrate(self, data: dict):
        data["ratelimit"] = {
            "params": {},
            "rate_limits": [],
            "whitelisted_address_pairs": [],
            "blacklisted_denoms": [],
            "pending_send_packet_sequence_numbers": [],
            "hour_epoch": {
                "epoch_number": "0",
                "duration": "3600s",
                "epoch_start_time": "0001-01-01T00:00:00Z",
                "epoch_start_height": "0"
            }
        }
