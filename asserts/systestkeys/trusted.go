// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

// Package systestkeys defines trusted assertions and keys to use in tests.
package systestkeys

import (
	"fmt"

	"github.com/snapcore/snapd/asserts"
)

const (
	TestRootPrivKey = `-----BEGIN PGP PRIVATE KEY BLOCK-----
Version: GnuPG v1

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
97WP6Yv2
=gJ0v
-----END PGP PRIVATE KEY BLOCK-----
`

	encodedTestRootAccount = `type: account
authority-id: testrootorg
account-id: testrootorg
display-name: Testrootorg
timestamp: 2018-06-27T14:25:40+02:00
username: testrootorg
validation: verified
sign-key-sha3-384: hIedp1AvrWlcDI4uS_qjoFLzjKl5enu4G2FYJpgB3Pj-tUzGlTQBxMBsBmi-tnJR

AcLBUgQAAQoABgUCWzOCRAAA31gQAE5QgyBuxF3DGlMP32+3G5soq0uDLKG+sqFIEj/8j1dwLG0u
ut7UPEf5iTZFDqqyAaFBRUPxx1cGB/6WFrks3X3/325hVzv5DYA9d4508BXdlNBA++t7tTdb4rU6
G57aVgbpMCdwdjRbRMv1LLVWnli1pj8Cvt/jiTMbJUwQ/CAO0UZA6EH+fAeHGB53NNvedAM1goWi
pS3XvruDtv8qTbVW9jNSIX1ADcLAbmM2xV2Vo54lfgN5NJd/4K4S7sPSsX7QLBghFkB0m9i5g/Qu
PecvJ9njebFF48yvG4W1owBNBfxD2oHNhK/GdtxsREDKgDXuIrhziXBzWNeYto8lCZ1D520k+xp+
2rL1TYSy9IixOzAf2qBhUTQdXsoVfmBOyExlYVQDIFO+X4ufbhLzy2pTE4KWvFvF58HzGradbix6
oUD5hiEjw1YoV8FKdMLDobcvGzgm+Kx/FQo2Iqm5GmfzPW/K3SntoptuHIDSk3B12F/F/EDoYiGS
MDWJJ4NMbFLMemJhvEI23IuZOTBEt27sGcgOju4wYkcsaHPEeXTGUBUgQADugBTwJtWmuybkmovM
aLn1kVYpht+0cZeAQR5L3nSOK7T3V+QvWSXt6PAiHJv+HnemrYarSmGVTDcpj1QuWXyX026RnkvP
SD73HCe5QPTjrvFvIa6o6n9khFgs
`

	TestRootKeyID = "hIedp1AvrWlcDI4uS_qjoFLzjKl5enu4G2FYJpgB3Pj-tUzGlTQBxMBsBmi-tnJR"

	encodedTestRootAccountKey = `type: account-key
authority-id: testrootorg
public-key-sha3-384: hIedp1AvrWlcDI4uS_qjoFLzjKl5enu4G2FYJpgB3Pj-tUzGlTQBxMBsBmi-tnJR
account-id: testrootorg
name: test-root
since: 2016-08-11T18:30:57+02:00
body-length: 717
sign-key-sha3-384: hIedp1AvrWlcDI4uS_qjoFLzjKl5enu4G2FYJpgB3Pj-tUzGlTQBxMBsBmi-tnJR

AcbBTQRWhcGAARAA8dC6HP+NfM5sNgCHH+bsQv4YLIR8glPfJ+HEXyaYdNO1+oFyX4nx7CpV5Umu
TYs7DPVpToAiN3snpBdPPKu5UEzkQ6OGDucf2bZnAInj7WzKwGnOA/Y/uQMduIyeFZ4mLnUNcF+M
e8LV0aS/pQhEdBUuRxEOi9zlv0p7X1bUs6LIUTubu6+smFtbdBBNOD+0qrvjf7CvsScrTsQswtvw
cLoB4GX94wK6RQrlkmYJPUFZqkdWt7cp0iq8d+Ts8UnT8sgWuFzkMCgBKritS7/545mE8AE0fsyF
Gt5+0jcjgs9LDk5gRO7EgoFLXPsEBdiLdVms7OGAwPGG00wfFYL3ho4PCfKq+mH0kOgUAynlJ7x8
MCR92eWEi/ylHXiO0jnRY8UsutrM76eLN41iUla/6j5DcsXxQB/xzlYkUdtXtYrn6L/DTsnixclu
3ogPzlPEFyVxv0vWIgkKLWXj2JRRt2uqe3K33TvdF0H+m6snZTStn7VY3if9fvyx14+tKh16ucdQ
a1zzJoTKTqYWX9B+ZfENGKJUnhTP0x7Cm6lg3EUGay/b5hsA4DBoqShuf/N0jVLojdhxi3Ck/DBN
lqCD0zy4uzvinjX+b4ay+LKBE3N15AsfEkWIwzI+1OdDlOWWqOxJkM6lrQ5hRQ1fHZoCiGjHbjeE
1RIFO2TAw2tpyUcAEQEAAQ==

AcLBXAQAAQoABgUCV8656QAKCRBMcZp594FxpNWlEADQgBlROdBTHpdZ3/9BbasxenUC3VXusMeK
0DmnsHrsAsyVk6xiHQQ3hWxvXKWoDkDsOhUqcQTsDBcIaZ18+qwpQciyItd+w3d7SSJ+MKSUpwsB
NOdgw1ykj7l1M/W7xAAPscFoV1xVSk9+rsLYFYDe23R+ecyotSmF+4QHj5b+hXeVIOUaqQTl5xPC
h0zVYNIUWv42q4Z+hiBS8+8UJ0G+7z/27XORkGHY6TXCt0aph7s5egr8Lm+/jq7c95HVsa7DwSpv
SqPajRnlyLiHFXUYAUPEU9oDgPwtLsqUkFfrv1WZ3ja1rDexgKBta+8BRyCAq3gPcMAjhiHXdjoW
90p893l9N6K82RiEOO9ic0pEezjQldg97oU+ajXNm3ryns+HX6hRd39rpzIsrbVdbCqun4RwMbCM
EVxgC/cuxMGcS40Co3O8wG3H/WIWOqcRQfolQTexmyzQljYt9WyWJdXmtPtaMzQGbOqE/dIjOK9j
xvrghVU4kX6fJFwPi+azMrluHV+WGSVxPCuLW8o2aipjOd1/bUQCL5OwRuaEWuLCiV01J8H/JjWV
hL4gGVqEM2KEPIDwY2yqX36jE7uN9O+mIPnS4Tdj0JQ5ZD1qh34wv+4QvhgNeyP120nuS1ykO9X0
A806uPC5QK1+cgRMUz8zJ0afDNwE/DvpBQvE5CIi9A==
`

	TestStorePrivKey = `-----BEGIN PGP PRIVATE KEY BLOCK-----
Version: GnuPG v1

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
o7ZSZ/h/bUY1EjE2
=Nc2M
-----END PGP PRIVATE KEY BLOCK-----
`

	TestStoreKeyID = "XCIC_Wvj9_hiAt0b10sDon74oGr3a6xGODkMZqrj63ZzNYUD5N87-ojjPoeN7f1Y"

	encodedTestStoreAccountKey = `type: account-key
authority-id: testrootorg
public-key-sha3-384: XCIC_Wvj9_hiAt0b10sDon74oGr3a6xGODkMZqrj63ZzNYUD5N87-ojjPoeN7f1Y
account-id: testrootorg
name: test-store
since: 2016-08-11T18:42:22+02:00
body-length: 717
sign-key-sha3-384: hIedp1AvrWlcDI4uS_qjoFLzjKl5enu4G2FYJpgB3Pj-tUzGlTQBxMBsBmi-tnJR

AcbBTQRWhcGAARAAmJqmZvsS58INTs+UQ+jfo836vBS5tkU/hk7c0huZe3So2gc9kaJvjkjhZ6g0
0/kGoidw3i2WkdMEp+JtvU9Ztfeu/Nn/OqkSc3Ap1KAmqL4OllPVII8H69w2zqvmo+PcqH0SvHAV
EoOC2ToXP0wEHnAZsbVu56AKrwpHDppPEIvaS6glrsEX1AXpOeMHZLVRtfsBB6dlLXuula1UrSAL
RFCEjXtqXqOto5Vo9C8p63lLBy19ifz4OriWAGBqZvFdItmo8VIPXSDdgMMHBlu2MSNJHVfo66yp
buos5Qs6PlVARACLJgzgplI5sDbzXtVx5O9Q8YJjz4NfU0WOWYPWvmANXDeMNixWoWvuixYvsXaG
x6mdD7Hh/gi8prkQmZ7gxW1MEOV9JThAqYjjs2ayGVD73EI2sKYxVwEg3iJToQ/cEz3O2U1HdmYj
QfRDJiX3GEPBXXttDrbPM42SHElouldmJ+PkJDLdkGmA85xYUoEKHdEFIkjFStQcyO5CkyNZN7SH
iAIwA4DpGMmFJ26maqVzJuiLvicri2FR/sJaSA24N8HbGne3gSS7WrSQS+jKe3IZPVy64NCoGvrW
o/HvTeqsIfihKPEpXm8QVtjNhtkVn3RdIUgOaNWyAfnZ4dW1TVIATe+OHDw2TNyImTjE0x75nL6B
1/Rrn+9VP9Swhv8AEQEAAQ==

AcLBXAQAAQoABgUCV866kwAKCRBMcZp594FxpHWHD/9AaZXqyT/Zsmq/VzmAMpd9JvCH4PHQKtAP
bXfP2Dnpa2wk2wuzQuSWunR8NDRyVh/aNVeTEZ9dFm/B8LR+U2O4rsHmFSeicmsTmo9u/HouRdEU
zeSc6cbAxMPpfNSjr5J+URLjGRT6oX5fEBmRPx/OC9pEIScMx7uKmTKEnuyMzLRNN/6HiGWKrFCo
nJdKkwRXrkCHyXWAOv1GumT7NDuyFcjAqt/UdHliTZkDBImKOsBmBVXMUjg7HCSS2uq/5WjStJ+B
JHQ4GSsXBvVINs6BncNWcvV6mCQ73D57MzGhqo997Zb4tSrn7UNGWK7GLCzV3e/pFlG7pw6HbgnQ
+rxU2Oj/TPVw0tcnUiRl2ttKpm+nua0Cl+MD+Gx0KXLAVp0ZGOQ9yGyP9AePFzcOR8SlRIgxi0EI
iJkSeYilqoKo3AJhnICRiqvAca2TGJoiJUryEgZ8jbTOElfaF2p+y0xvXGlWbKZm1gzGyvFM5fV5
hJTlp/am+2uVn6U8wPACir4PrbuXYo7L4MIXww2OEO0ruBIaLARbc5IutSWmw6AEYQUxtsa9bdHV
Zin7LGbEj6lZm8GycWQwh4B6Vnt6dJRIyPc/9G7uM8Ds/2Wa7+yAxhiPqm8DwlbOYh1npw4X4TLD
IMGnTv5N3zllI+Xz4rqJzNTzEbvOIcrqWxCedQe79A==
`
)

var (
	TestRootAccount    asserts.Assertion
	TestRootAccountKey asserts.Assertion
	// here for convenience, does not need to be in the trusted set
	TestStoreAccountKey asserts.Assertion
	// Testing-only trusted assertions for injecting in the the system trusted set.
	Trusted []asserts.Assertion
)

func init() {
	acct, err := asserts.Decode([]byte(encodedTestRootAccount))
	if err != nil {
		panic(fmt.Sprintf("cannot decode trusted assertion: %v", err))
	}
	accKey, err := asserts.Decode([]byte(encodedTestRootAccountKey))
	if err != nil {
		panic(fmt.Sprintf("cannot decode trusted assertion: %v", err))
	}
	storeAccKey, err := asserts.Decode([]byte(encodedTestStoreAccountKey))
	if err != nil {
		panic(fmt.Sprintf("cannot decode test store assertion: %v", err))
	}

	TestRootAccount = acct
	TestRootAccountKey = accKey
	TestStoreAccountKey = storeAccKey
	Trusted = []asserts.Assertion{TestRootAccount, TestRootAccountKey}
}
