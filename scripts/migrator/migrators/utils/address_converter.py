from eth_utils import is_hex_address, to_checksum_address, keccak
from bech32 import bech32_decode, convertbits, bech32_encode

# Converts a bech32 to hex
def bech32_to_hex(addr: str, bech32_prefix: str = "cosmos") -> str:
    # Check if it starts with the prefix
    if addr.startswith(bech32_prefix):
        hrp, data = bech32_decode(addr)
        if hrp != bech32_prefix or data is None:
            raise ValueError("Must provide a valid Bech32 address")
        # Convert 5-bit words to 8-bit bytes
        decoded_bytes = bytes(convertbits(data, 5, 8, False))
        if len(decoded_bytes) != 20:
            raise ValueError("Bech32 address decoded to invalid length")
        return to_checksum_address(decoded_bytes)

    # If it's a hex we return the hex
    if not addr.startswith("0x"):
        addr = "0x" + addr

    # If its neither
    if not is_hex_address(addr):
        raise ValueError(f"{addr} is not a valid Ethereum or Cosmos address")

    return to_checksum_address(addr)

# Converts a hex to bech32
def hex_to_bech32(hex_addr: str, bech32_prefix: str = "cosmos") -> str:
    # Check the beginning
    if not hex_addr.startswith("0x"):
        raise ValueError("Hex address must start with 0x")

    # Check if it's a hex address
    if not is_hex_address(hex_addr):
        raise ValueError(f"{hex_addr} is not a valid hex address")

    addr_bytes = bytes.fromhex(hex_addr[2:])  # remove 0x
    if len(addr_bytes) != 20:
        raise ValueError("Ethereum address must be 20 bytes")

    data = convertbits(addr_bytes, 8, 5)
    return bech32_encode(bech32_prefix, data)

# Converts a address between different prefixes
def convert_bech32_prefix(addr: str, bech32_prefix_out: str = "cosmos"):
    # Convert to bytes
    _, data = bech32_decode(addr)

    # Convert back
    return bech32_encode(bech32_prefix_out, data)