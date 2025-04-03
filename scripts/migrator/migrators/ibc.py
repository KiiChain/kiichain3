from migrators import Migrator


class IBC(Migrator):
    # Set the param as default
    def migrate(self, data: dict):
        data["ibc"] = {
            "client_genesis": {
                "clients": [
                    {
                        "client_id": "09-localhost",
                        "client_state": {
                            "@type": "/ibc.lightclients.localhost.v2.ClientState",
                            "latest_height": {
                                "revision_number": "1",
                                "revision_height": "28",
                            },
                        },
                    }
                ],
                "clients_consensus": [],
                "clients_metadata": [],
                "params": {"allowed_clients": ["*"]},
                "create_localhost": False,
                "next_client_sequence": "0",
            },
            "connection_genesis": {
                "connections": [
                    {
                        "id": "connection-localhost",
                        "client_id": "09-localhost",
                        "versions": [
                            {
                                "identifier": "1",
                                "features": ["ORDER_ORDERED", "ORDER_UNORDERED"],
                            }
                        ],
                        "state": "STATE_OPEN",
                        "counterparty": {
                            "client_id": "09-localhost",
                            "connection_id": "connection-localhost",
                            "prefix": {"key_prefix": "aWJj"},
                        },
                        "delay_period": "0",
                    }
                ],
                "client_connection_paths": [],
                "next_connection_sequence": "0",
                "params": {"max_expected_time_per_block": "30000000000"},
            },
            "channel_genesis": {
                "channels": [],
                "acknowledgements": [],
                "commitments": [],
                "receipts": [],
                "send_sequences": [],
                "recv_sequences": [],
                "ack_sequences": [],
                "next_channel_sequence": "0",
                "params": {
                    "upgrade_timeout": {
                        "height": {"revision_number": "0", "revision_height": "0"},
                        "timestamp": "600000000000",
                    }
                },
            },
        }
