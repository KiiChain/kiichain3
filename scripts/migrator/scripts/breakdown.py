import json

with open("export_testnet_jq.json", "r") as f:
    export_data = json.loads(f.read())

# Iterate all the keys
for key, value in export_data['app_state'].items():
    with open(f"breakdown/{key}.json", "w") as f:
        f.write(json.dumps({key: value}, indent=2))
