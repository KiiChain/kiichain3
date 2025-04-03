python scripts/migrator/main.py export.json export_out.json
jq -S . export_out.json > export_sorted.json
evmd genesis validate /home/korok/kii/kiichain3/export_sorted.json
