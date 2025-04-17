import json

with open("testnet_data/export_testnet.json", "r") as file:
    export_data = json.loads(file.read())

bank_balances = export_data["app_state"]["bank"]["balances"]

address_balance_map = {}

for balance in bank_balances:
    for coin in balance["coins"]:
        if coin["denom"] == "ukii":
            address_balance_map[balance["address"]] = int(coin["amount"])

sorted_items = sorted(address_balance_map.items(), key=lambda item: item[1], reverse=True)

top_100 = sorted_items[:100]

auth_accounts = export_data["app_state"]["auth"]["accounts"]

address_account_number_map = {}

for account in auth_accounts:
    if "base_account" in account.keys():
        address_account_number_map[account['base_account']['address']] = account['base_account']['account_number']
    else:
        address_account_number_map[account['address']] = account['account_number']


for wallet, balance in top_100:
    account_number = address_account_number_map.get(wallet)
    formatted = f"{int(balance) / 1_000_000:,.6f}"
    print(f"Wallet: {wallet}, Balance: {formatted} kii, Account number: {account_number}")
