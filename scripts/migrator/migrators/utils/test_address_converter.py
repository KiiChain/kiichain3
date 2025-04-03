import unittest

from address_converter import bech32_to_hex, hex_to_bech32, convert_bech32_prefix

class TestCliUtils(unittest.TestCase):
    def test_account_to_hex_case_kii_1(self):
        addr = "kii1grewylu4k58y6yuyn5yay5zqqj3d2l6skxu8xc"
        expected_eth = "0x40f2e27f95b50E4D13849d09d2504004A2D57f50"
        self.assertEqual(bech32_to_hex(addr, "kii"), expected_eth)

    def test_account_to_hex_case_kii_2(self):
        addr = "kii19xtgpm6scmwcjazsza42eyeapzm4dh0xax6trr"
        expected = "0x299680Ef50C6DD897450176aac933D08B756DDe6"
        self.assertEqual(bech32_to_hex(addr, "kii"), expected)

    def test_account_to_hex_case_hex_1(self):
        addr = "0x7979f3A3c912D73c953fb59de0311bfA9b176730"
        expected_eth = "kii109ul8g7fzttne9flkkw7qvgml2d3weesatzpa0"
        self.assertEqual(hex_to_bech32(addr, "kii"), expected_eth)

    def test_account_to_hex_case_hex_2(self):
        addr = "0xA18344d76Cf42dB408db7f27d1167BaeBeDFe1Ee"
        expected = "kii15xp5f4mv7skmgzxm0unaz9nm46ldlc0w93d8qa"
        self.assertEqual(hex_to_bech32(addr, "kii"), expected)

    def test_account_to_hex_case_kiicons_1(self):
        addr = "kiivalconspub1grewylu4k58y6yuyn5yay5zqqj3d2l6s4pdjqz"
        expected_eth = "0x40f2e27f95b50E4D13849d09d2504004A2D57f50"
        self.assertEqual(bech32_to_hex(addr, "kiivalconspub"), expected_eth)

    def test_account_to_hex_case_kiicons_2(self):
        addr = "kiivalconspub19xtgpm6scmwcjazsza42eyeapzm4dh0x7pt79e"
        expected = "0x299680Ef50C6DD897450176aac933D08B756DDe6"
        self.assertEqual(bech32_to_hex(addr, "kiivalconspub"), expected)

    def test_convert_bech32_prefix_1(self):
        addr = "kiivaloper1px0cafksu5akmx4dwmqcskhjczgp72n28reet3"
        expected = "kii1px0cafksu5akmx4dwmqcskhjczgp72n2j4z229"
        self.assertEqual(convert_bech32_prefix(addr, "kii"), expected)

if __name__ == '__main__':
    unittest.main()
