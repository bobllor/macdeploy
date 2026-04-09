import system.utils as utils

def test_filevault_successful_match():
    # NOTE: these are randomly generated
    keys: list[str] = [
        "A7F3-K9X2-L4Q8-M2ZT-P8R5-W1VN",
        "Z4D8-H2MP-9XQ3-T7LS-C6YB-R5KF",
        "Q9W2-J8VL-X3A7-N6RT-P4ZK-M1DH",
        "M5XK-2P9T-L8Q3-V7RB-Y4ZN-H6CW",
        "T3Z8-Y6QP-W2LX-N5KD-R9M4-J7BV",
        "K2N7-R4YW-Z8PT-M3QH-X6LC-V9DA",
        "P8M4-J2DZ-V7RT-Y3XN-L5QK-W6HB",
        "X6R3-L9PT-K2MW-N7ZB-D4Q8-Y5VH",
        "H4Q7-X2LC-T9RM-P6ZW-K3ND-V8YJ",
        "W9Z5-N3KT-Y7MP-L2RX-Q8VH-D4CB",
    ]

    for key in keys:
        assert utils.is_filevault_key(key)

def test_filevault_fail_match():
    keys: list[str] = [
        "A7F3-K9X2-L4Q8M2ZTP8R5-W1VN",
        "Z4D8H2MP9XQ3T7LSC6YBR5KF",
        "Q9W2-J8VL-X3A7-N6RT-ZK-M1DH",
        "M5XK-2P9T-L8Q3-Y4ZN-H6CW",
        "T3Z8-Y6QP-W2LX-N5KDV",
        "K2N7-R4YW-Z8PT-M3QH-X6LC-V9DAZZ",
        "P8M4-J2DZ-V7RTK-W6HB",
        "X6R3-L9PT-K2MW-Y5VH",
        "thisisnotevenakey.txt",
        "W9Z5-N3KT-Y7MP-L2RX-Q8VH",
    ]

    for key in keys:
        assert not utils.is_filevault_key(key)