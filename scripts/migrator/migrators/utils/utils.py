import json
import base64
from decimal import Decimal, getcontext

getcontext().prec = 50

def replace_in_dict(obj, replace_map):
    if isinstance(obj, dict):
        return {
            key: replace_in_dict(value, replace_map)
            for key, value in obj.items()
        }
    elif isinstance(obj, list):
        return [replace_in_dict(item, replace_map) for item in obj]
    elif isinstance(obj, str):
        return replace_map.get(obj, obj)  # only replace exact matches
    else:
        return obj

# Do replace in a json in a serialized form
def replace_in_serialized(data: dict, replacement_dict: dict[str, str]) -> dict:
    json_str = json.dumps(data)

    # Iterate all the replacement dict
    for key, value in replacement_dict.items():
        json_str = json_str.replace(key, value)

    # Return the final data
    return json.loads(json_str)

# Migrate a list of coins to 18 decimals for a selected denom
def coins_to_decimals(coins: list[dict], denom: str = "ukii"):
    for coin in coins:
        coin_denom = coin['denom']
        amount = coin['amount']

        # Check if ukii
        if coin_denom == denom:
            coin['amount'] = add_decimals(amount)

# Migrate a list of dec coins to have 18 decimals for a selected denom
def dec_coins_to_decimals(dec_coins: list[dict], denom: str = "ukii"):
    for dec_coin in dec_coins:
        coin_denom = dec_coin['denom']
        amount = dec_coin['amount']

        # Check if ukii
        if coin_denom == denom:
            dec_coin['amount'] = f"{dec_add_decimals(amount)}"

# Turns a base64 string into hex
def base64_to_hex(string: str):
    return base64.b64decode(string).hex()

# Add 12 more decimals to a string
def dec_add_decimals(number: str, decimals: int = 12) -> str:
    value = Decimal(number)
    shifted = value * Decimal(f"1e{decimals}")
    return f"{shifted}"

# Add 12 mode decimals to a number as a string
def add_decimals(number: str, decimals: int = 12) -> str:
    wei_decimals = "".zfill(decimals)
    return f"{number}{wei_decimals}"