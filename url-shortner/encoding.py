
# a-z | A - Z | 0 - 9
# 26 + 26 + 10 = 62 ~ 6 bits

map = {}

chars = [chr(i) for i in range(ord('a'), ord("z") + 1)]
chars += [chr(i) for i in range(ord('A'), ord("Z") + 1)]
chars += [chr(i) for i in range(ord('0'), ord("9") + 1)]

map.update({i: chars[i] for i in range(len(chars))})


def encode(code: int) -> str:
    """
    Encode an integer to a string.
    79 -> 1001111
    1001111 -> 000001 001111 # left pad zereos to multiple of 6
    -> b p
    """
    bits = bin(code)[2:]
    pad = (-len(bits)) % 6
    bits = bits.zfill(len(bits) + pad)
    result = ''
    for i in range(0, len(bits), 6):
        chunk = bits[i:i+6]
        idx = int(chunk, 2)
        result += map[idx]
    return result
