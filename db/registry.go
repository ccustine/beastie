package db

import (
	"bytes"
	"github.com/hashicorp/go-msgpack/codec"
)

const (
	AIRCRAFT_PREFIX     = "AC-"
	REGISTRATION_PREFIX = "RG-"
)

var (
	regToCountry = map[string]string{
		"3A":   "Monaco",
		"3B":   "Mauritius",
		"3C":   "Equatorial Guinea",
		"3D":   "Swaziland",
		"3X":   "Guinea",
		"4K":   "Azerbaijan",
		"4L":   "Georgia",
		"4R":   "Sri Lanka",
		"4W":   "Yemen",
		"4X":   "Israel",
		"5A":   "Libyan Arab Jamahiriya",
		"5B":   "Cyprus",
		"5H":   "United Republic of Tanzania",
		"5N":   "Nigeria",
		"5R":   "Madagascar",
		"5T":   "Mauritania",
		"5U":   "Niger",
		"5V":   "Togo",
		"5W":   "Samoa",
		"5X":   "Uganda",
		"5Y":   "Kenya",
		"6O":   "Somalia",
		"6V":   "Senegal",
		"6W":   "Senegal",
		"6Y":   "Jamaica",
		"7P":   "Lesotho",
		"7QY":  "Malawi",
		"7T":   "Algeria",
		"8P":   "Barbados",
		"8Q":   "Maldives",
		"8R":   "Guyana",
		"9A":   "Croatia",
		"9G":   "Ghana",
		"9H":   "Malta",
		"9J":   "Zambia",
		"9K":   "Kuwait",
		"9L":   "Sierra Leone",
		"9M":   "Malaysia",
		"9N":   "Nepal",
		"9Q":   "Zaire",
		"9U":   "Burundi",
		"9V":   "Singapore",
		"9XR":  "Rwanda",
		"9Y":   "Trinidad and Tobago",
		"A2":   "Botswana",
		"A3":   "Tonga",
		"A40":  "Oman",
		"A5":   "Bhutan",
		"A6":   "United Arab Emirates",
		"A7":   "Qatar",
		"A9C":  "Bahrain",
		"AP":   "Pakistan",
		"B":    "China",
		"C,CF": "Canada",
		"C2":   "Nauru",
		"C5":   "Gambia",
		"C6":   "Bahamas",
		"C9":   "Mozambique",
		"CC":   "Chile",
		"CCCP": "Union of Soviet Socialist Republics",
		"CN":   "Morocco",
		"CP":   "Bolivia",
		"CR":   "Portugal",
		"CS":   "Portugal",
		"CU":   "Cuba",
		"CX":   "Uruguay",
		"D":    "Germany",
		"D2":   "Angola",
		"D4":   "Cape Verde",
		"DQ":   "Fiji",
		"EC":   "Spain",
		"EI":   "Ireland",
		"EJ":   "Ireland",
		"EK":   "Armenia",
		"EL":   "Liberia",
		"EP":   "Iran",
		"ER":   "Moldova",
		"ES":   "Estonia",
		"ET":   "Ethiopia",
		"EW":   "Belarus",
		"EX":   "Kyrgyzstan",
		"EY":   "Tajikistan",
		"EZ":   "Turkmenistan",
		"F":    "France",
		"G":    "United Kingdom",
		"H4":   "Solomon Islands",
		"HA":   "Hungary",
		"HB":   "Switzerland or Liechtenstein",
		"HC":   "Ecuador",
		"HH":   "Haiti",
		"HI":   "Dominican Republic",
		"HK":   "Columbia",
		"HL":   "Republic of Korea",
		"HP":   "Panama",
		"HR":   "Honduras",
		"HS":   "Thailand",
		"HZ":   "Saudi Arabia",
		"I":    "Italy",
		"J2":   "Djibouti",
		"J3":   "Grenada",
		"J5":   "Guinea Bissau",
		"J6":   "Saint Lucia",
		"J7":   "Dominica",
		"J8":   "Saint Vincent and the Grenadines",
		"JA":   "Japan",
		"JY":   "Jordan",
		"LN":   "Norway",
		"LQ":   "Argentina",
		"LV":   "Argentina",
		"LX":   "Luxembourg",
		"LY":   "Lithuania",
		"LZ":   "Bulgaria",
		"N":    "United States",
		"OB":   "Peru",
		"OD":   "Lebanon",
		"OE":   "Austria",
		"OH":   "Finland",
		"OK":   "Czech Republic",
		"OM":   "Slovakia",
		"OO":   "Belgium",
		"OY":   "Denmark",
		"P":    "Democratic People's Republic of Korea",
		"P2":   "Papua New Guinea",
		"P4":   "Aruba",
		"PH":   "Netherlands, Kingdom of the",
		"PJ":   "Netherlands, Netherlands Antilles",
		"PK":   "Indonesia or West Iran",
		"PP":   "Brazil",
		"PR":   "Brazil",
		"PT":   "Brazil",
		"PU":   "Brazil",
		"PZ":   "Suriname",
		"RA":   "Russian Federation",
		"RDPL": "Lao People's Democratic Republic",
		"RP":   "Philippines",
		"S2":   "Bangladesh",
		"S7":   "Seychelles",
		"S9":   "Sao Tome and Principe",
		"SE":   "Sweden",
		"SP":   "Poland",
		"ST":   "Sudan",
		"SU":   "Egypt, Arab Republic of",
		"SX":   "Greece",
		"T7":   "San Marino",
		"T9":   "Bosnia and Herzegovina",
		"TC":   "Turkey",
		"TF":   "Iceland",
		"TG":   "Guatemala",
		"TI":   "Costa Rica",
		"TJ":   "Cameroon",
		"TL":   "Central African Republic",
		"TN":   "Congo",
		"TR":   "Gabon",
		"TS":   "Tunisia",
		"TT":   "Chad",
		"TU":   "Ivory Coast",
		"TY":   "Benin",
		"TZ":   "Mali",
		"UK":   "Uzbekistan",
		"UN":   "Kazakhstan",
		"UR":   "Ukraine",
		"V2":   "Antigua and Barbuda",
		"V3":   "Belize",
		"V4":   "Saint Kitts and Nevis",
		"V5":   "Namibia",
		"V6":   "Micronesia, Federated States of",
		"V7":   "Marshall Islands",
		"V8":   "Brunei Darussalam",
		"VH":   "Australia",
		"VP":   "United Kingdom, Colonies and Protectorates",
		"VQ":   "United Kingdom, Colonies and Protectorates",
		"VR":   "United Kingdom, Colonies and Protectorates",
		"VT":   "India",
		"XA":   "Mexico",
		"XB":   "Mexico",
		"XC":   "Mexico",
		"XT":   "Burkina Faso",
		"XU":   "Cambodia",
		"XV":   "Vietnam",
		"YA":   "Afghanistan",
		"YI":   "Iraq",
		"YJ":   "Vanuatu",
		"YK":   "Syrian Arab Republic",
		"YL":   "Latvia",
		"YN":   "Nicaragua",
		"YR":   "Romania",
		"YS":   "El Salvador",
		"YU":   "Yugoslavia",
		"YV":   "Venezuela",
		"Z":    "Zimbabwe",
		"Z3":   "Macedonia",
		"ZK":   "New Zealand",
		"ZL":   "New Zealand",
		"ZM":   "New Zealand",
		"ZP":   "Paraguay",
		"ZS":   "South Africa",
		"ZT":   "South Africa",
		"ZU":   "South Africa",
	}

)

func IsMil(icao uint32) bool {
	switch {
	case icao >= uint32(0xADF7C8) && icao <= uint32(0xAFFFFF):
		return true
	default:
		return false
	}
}

func IcaoToCountry(icao uint32) string {
	switch {
	case icao >= uint32(0x000000) && icao <= uint32(0x003FFF):
		return "(unallocated)"
	case icao >= uint32(0x004000) && icao <= uint32(0x0043FF):
		return "Zimbabwe"
	case icao >= uint32(0x006000) && icao <= uint32(0x006FFF):
		return "Mozambique"
	case icao >= uint32(0x008000) && icao <= uint32(0x00FFFF):
		return "South Africa"
	case icao >= uint32(0x010000) && icao <= uint32(0x017FFF):
		return "Egypt"
	case icao >= uint32(0x018000) && icao <= uint32(0x01FFFF):
		return "Libya"
	case icao >= uint32(0x020000) && icao <= uint32(0x027FFF):
		return "Morocco"
	case icao >= uint32(0x028000) && icao <= uint32(0x02FFFF):
		return "Tunisia"
	case icao >= uint32(0x030000) && icao <= uint32(0x0303FF):
		return "Botswana"
	case icao >= uint32(0x032000) && icao <= uint32(0x032FFF):
		return "Burundi"
	case icao >= uint32(0x034000) && icao <= uint32(0x034FFF):
		return "Cameroon"
	case icao >= uint32(0x035000) && icao <= uint32(0x0353FF):
		return "Comoros"
	case icao >= uint32(0x036000) && icao <= uint32(0x036FFF):
		return "Congo"
	case icao >= uint32(0x038000) && icao <= uint32(0x038FFF):
		return "Cote d Ivoire"
	case icao >= uint32(0x03E000) && icao <= uint32(0x03EFFF):
		return "Gabon"
	case icao >= uint32(0x040000) && icao <= uint32(0x040FFF):
		return "Ethiopia"
	case icao >= uint32(0x042000) && icao <= uint32(0x042FFF):
		return "Equatorial Guinea"
	case icao >= uint32(0x044000) && icao <= uint32(0x044FFF):
		return "Ghana"
	case icao >= uint32(0x046000) && icao <= uint32(0x046FFF):
		return "Guinea"
	case icao >= uint32(0x04A000) && icao <= uint32(0x04A3FF):
		return "Lesotho"
	case icao >= uint32(0x04C000) && icao <= uint32(0x04CFFF):
		return "Kenya"
	case icao >= uint32(0x050000) && icao <= uint32(0x050FFF):
		return "Liberia"
	case icao >= uint32(0x054000) && icao <= uint32(0x054FFF):
		return "Madagascar"
	case icao >= uint32(0x058000) && icao <= uint32(0x058FFF):
		return "Malawi"
	case icao >= uint32(0x05A000) && icao <= uint32(0x05A3FF):
		return "Maldives"
	case icao >= uint32(0x05C000) && icao <= uint32(0x05CFFF):
		return "Mali"
	case icao >= uint32(0x05E000) && icao <= uint32(0x05E3FF):
		return "Mauritania"
	case icao >= uint32(0x060000) && icao <= uint32(0x0603FF):
		return "Mauritius"
	case icao >= uint32(0x062000) && icao <= uint32(0x062FFF):
		return "Niger"
	case icao >= uint32(0x064000) && icao <= uint32(0x064FFF):
		return "Nigeria"
	case icao >= uint32(0x068000) && icao <= uint32(0x068FFF):
		return "Uganda"
	case icao >= uint32(0x06A000) && icao <= uint32(0x06A3FF):
		return "Qatar"
	case icao >= uint32(0x06C000) && icao <= uint32(0x06CFFF):
		return "Central African Republic"
	case icao >= uint32(0x06E000) && icao <= uint32(0x06EFFF):
		return "Rwanda"
	case icao >= uint32(0x070000) && icao <= uint32(0x070FFF):
		return "Senegal"
	case icao >= uint32(0x074000) && icao <= uint32(0x0743FF):
		return "Seychelles"
	case icao >= uint32(0x076000) && icao <= uint32(0x0763FF):
		return "Sierra Leone"
	case icao >= uint32(0x078000) && icao <= uint32(0x078FFF):
		return "Somalia"
	case icao >= uint32(0x07A000) && icao <= uint32(0x07A3FF):
		return "Swaziland"
	case icao >= uint32(0x07C000) && icao <= uint32(0x07CFFF):
		return "Sudan"
	case icao >= uint32(0x080000) && icao <= uint32(0x080FFF):
		return "Tanzania"
	case icao >= uint32(0x084000) && icao <= uint32(0x084FFF):
		return "Chad"
	case icao >= uint32(0x088000) && icao <= uint32(0x088FFF):
		return "Togo"
	case icao >= uint32(0x08A000) && icao <= uint32(0x08AFFF):
		return "Zambia"
	case icao >= uint32(0x08C000) && icao <= uint32(0x08CFFF):
		return "D R Congo"
	case icao >= uint32(0x090000) && icao <= uint32(0x090FFF):
		return "Angola"
	case icao >= uint32(0x094000) && icao <= uint32(0x0943FF):
		return "Benin"
	case icao >= uint32(0x096000) && icao <= uint32(0x0963FF):
		return "Cape Verde"
	case icao >= uint32(0x098000) && icao <= uint32(0x0983FF):
		return "Djibouti"
	case icao >= uint32(0x098000) && icao <= uint32(0x0983FF):
		return "Djibouti"
	case icao >= uint32(0x09A000) && icao <= uint32(0x09AFFF):
		return "Gambia"
	case icao >= uint32(0x09C000) && icao <= uint32(0x09CFFF):
		return "Burkina Faso"
	case icao >= uint32(0x09E000) && icao <= uint32(0x09E3FF):
		return "Sao Tome"
	case icao >= uint32(0x0A0000) && icao <= uint32(0x0A7FFF):
		return "Algeria"
	case icao >= uint32(0x0A8000) && icao <= uint32(0x0A8FFF):
		return "Bahamas"
	case icao >= uint32(0x0AA000) && icao <= uint32(0x0AA3FF):
		return "Barbados"
	case icao >= uint32(0x0AB000) && icao <= uint32(0x0AB3FF):
		return "Belize"
	case icao >= uint32(0x0AC000) && icao <= uint32(0x0ACFFF):
		return "Colombia"
	case icao >= uint32(0x0AE000) && icao <= uint32(0x0AEFFF):
		return "Costa Rica"
	case icao >= uint32(0x0B0000) && icao <= uint32(0x0B0FFF):
		return "Cuba"
	case icao >= uint32(0x0B2000) && icao <= uint32(0x0B2FFF):
		return "El Salvador"
	case icao >= uint32(0x0B4000) && icao <= uint32(0x0B4FFF):
		return "Guatemala"
	case icao >= uint32(0x0B6000) && icao <= uint32(0x0B6FFF):
		return "Guyana"
	case icao >= uint32(0x0B8000) && icao <= uint32(0x0B8FFF):
		return "Haiti"
	case icao >= uint32(0x0BA000) && icao <= uint32(0x0BAFFF):
		return "Honduras"
	case icao >= uint32(0x0BC000) && icao <= uint32(0x0BC3FF):
		return "St.Vincent + Grenadines"
	case icao >= uint32(0x0BE000) && icao <= uint32(0x0BEFFF):
		return "Jamaica"
	case icao >= uint32(0x0C0000) && icao <= uint32(0x0C0FFF):
		return "Nicaragua"
	case icao >= uint32(0x0C2000) && icao <= uint32(0x0C2FFF):
		return "Panama"
	case icao >= uint32(0x0C4000) && icao <= uint32(0x0C4FFF):
		return "Dominican Republic"
	case icao >= uint32(0x0C6000) && icao <= uint32(0x0C6FFF):
		return "Trinidad and Tobago"
	case icao >= uint32(0x0C8000) && icao <= uint32(0x0C8FFF):
		return "Suriname"
	case icao >= uint32(0x0CA000) && icao <= uint32(0x0CA3FF):
		return "Antigua + Barbuda"
	case icao >= uint32(0x0CC000) && icao <= uint32(0x0CC3FF):
		return "Grenada"
	case icao >= uint32(0x0D0000) && icao <= uint32(0x0D7FFF):
		return "Mexico"
	case icao >= uint32(0x0D8000) && icao <= uint32(0x0DFFFF):
		return "Venezuela"
	case icao >= uint32(0x100000) && icao <= uint32(0x1FFFFF):
		return "Russia"
	case icao >= uint32(0x200000) && icao <= uint32(0x27FFFF):
		return "(reserved, AFI)"
	case icao >= uint32(0x201000) && icao <= uint32(0x2013FF):
		return "Namibia"
	case icao >= uint32(0x202000) && icao <= uint32(0x2023FF):
		return "Eritrea"
	case icao >= uint32(0x280000) && icao <= uint32(0x2FFFFF):
		return "(reserved, SAM)"
	case icao >= uint32(0x300000) && icao <= uint32(0x33FFFF):
		return "Italy"
	case icao >= uint32(0x340000) && icao <= uint32(0x37FFFF):
		return "Spain"
	case icao >= uint32(0x380000) && icao <= uint32(0x3BFFFF):
		return "France"
	case icao >= uint32(0x3C0000) && icao <= uint32(0x3FFFFF):
		return "Germany"
	case icao >= uint32(0x400000) && icao <= uint32(0x43FFFF):
		return "United Kingdom"
	case icao >= uint32(0x440000) && icao <= uint32(0x447FFF):
		return "Austria"
	case icao >= uint32(0x448000) && icao <= uint32(0x44FFFF):
		return "Belgium"
	case icao >= uint32(0x450000) && icao <= uint32(0x457FFF):
		return "Bulgaria"
	case icao >= uint32(0x458000) && icao <= uint32(0x45FFFF):
		return "Denmark"
	case icao >= uint32(0x460000) && icao <= uint32(0x467FFF):
		return "Finland"
	case icao >= uint32(0x468000) && icao <= uint32(0x46FFFF):
		return "Greece"
	case icao >= uint32(0x470000) && icao <= uint32(0x477FFF):
		return "Hungary"
	case icao >= uint32(0x478000) && icao <= uint32(0x47FFFF):
		return "Norway"
	case icao >= uint32(0x480000) && icao <= uint32(0x487FFF):
		return "Netherlands"
	case icao >= uint32(0x488000) && icao <= uint32(0x48FFFF):
		return "Poland"
	case icao >= uint32(0x490000) && icao <= uint32(0x497FFF):
		return "Portugal"
	case icao >= uint32(0x498000) && icao <= uint32(0x49FFFF):
		return "Czech Republic"
	case icao >= uint32(0x4A0000) && icao <= uint32(0x4A7FFF):
		return "Romania"
	case icao >= uint32(0x4A8000) && icao <= uint32(0x4AFFFF):
		return "Sweden"
	case icao >= uint32(0x4B0000) && icao <= uint32(0x4B7FFF):
		return "Switzerland"
	case icao >= uint32(0x4B8000) && icao <= uint32(0x4BFFFF):
		return "Turkey"
	case icao >= uint32(0x4C0000) && icao <= uint32(0x4C7FFF):
		return "Yugoslavia"
	case icao >= uint32(0x4C8000) && icao <= uint32(0x4C83FF):
		return "Cyprus"
	case icao >= uint32(0x4CA000) && icao <= uint32(0x4CAFFF):
		return "Ireland"
	case icao >= uint32(0x4CC000) && icao <= uint32(0x4CCFFF):
		return "Iceland"
	case icao >= uint32(0x4D0000) && icao <= uint32(0x4D03FF):
		return "Luxembourg"
	case icao >= uint32(0x4D2000) && icao <= uint32(0x4D23FF):
		return "Malta"
	case icao >= uint32(0x4D4000) && icao <= uint32(0x4D43FF):
		return "Monaco"
	case icao >= uint32(0x500000) && icao <= uint32(0x5003FF):
		return "San Marino"
	case icao >= uint32(0x500000) && icao <= uint32(0x5FFFFF):
		return "(reserved, EUR/NAT)"
	case icao >= uint32(0x501000) && icao <= uint32(0x5013FF):
		return "Albania"
	case icao >= uint32(0x501C00) && icao <= uint32(0x501FFF):
		return "Croatia"
	case icao >= uint32(0x502C00) && icao <= uint32(0x502FFF):
		return "Latvia"
	case icao >= uint32(0x503C00) && icao <= uint32(0x503FFF):
		return "Lithuania"
	case icao >= uint32(0x504C00) && icao <= uint32(0x504FFF):
		return "Moldova"
	case icao >= uint32(0x505C00) && icao <= uint32(0x505FFF):
		return "Slovakia"
	case icao >= uint32(0x506C00) && icao <= uint32(0x506FFF):
		return "Slovenia"
	case icao >= uint32(0x507C00) && icao <= uint32(0x507FFF):
		return "Uzbekistan"
	case icao >= uint32(0x508000) && icao <= uint32(0x50FFFF):
		return "Ukraine"
	case icao >= uint32(0x510000) && icao <= uint32(0x5103FF):
		return "Belarus"
	case icao >= uint32(0x511000) && icao <= uint32(0x5113FF):
		return "Estonia"
	case icao >= uint32(0x512000) && icao <= uint32(0x5123FF):
		return "Macedonia"
	case icao >= uint32(0x513000) && icao <= uint32(0x5133FF):
		return "Bosnia + Herzegovina"
	case icao >= uint32(0x514000) && icao <= uint32(0x5143FF):
		return "Georgia"
	case icao >= uint32(0x515000) && icao <= uint32(0x5153FF):
		return "Tajikistan"
	case icao >= uint32(0x600000) && icao <= uint32(0x6003FF):
		return "Armenia"
	case icao >= uint32(0x600000) && icao <= uint32(0x67FFFF):
		return "(reserved, MID)"
	case icao >= uint32(0x600800) && icao <= uint32(0x600BFF):
		return "Azerbaijan"
	case icao >= uint32(0x601000) && icao <= uint32(0x6013FF):
		return "Kyrgyzstan"
	case icao >= uint32(0x601800) && icao <= uint32(0x601BFF):
		return "Turkmenistan"
	case icao >= uint32(0x680000) && icao <= uint32(0x6FFFFF):
		return "(reserved, ASIA)"
	case icao >= uint32(0x680000) && icao <= uint32(0x6803FF):
		return "Bhutan"
	case icao >= uint32(0x681000) && icao <= uint32(0x6813FF):
		return "Micronesia"
	case icao >= uint32(0x682000) && icao <= uint32(0x6823FF):
		return "Mongolia"
	case icao >= uint32(0x683000) && icao <= uint32(0x6833FF):
		return "Kazakhstan"
	case icao >= uint32(0x684000) && icao <= uint32(0x6843FF):
		return "Palau"
	case icao >= uint32(0x700000) && icao <= uint32(0x700FFF):
		return "Afghanistan"
	case icao >= uint32(0x702000) && icao <= uint32(0x702FFF):
		return "Bangladesh"
	case icao >= uint32(0x704000) && icao <= uint32(0x704FFF):
		return "Myanmar"
	case icao >= uint32(0x706000) && icao <= uint32(0x706FFF):
		return "Kuwait"
	case icao >= uint32(0x708000) && icao <= uint32(0x708FFF):
		return "Laos"
	case icao >= uint32(0x70A000) && icao <= uint32(0x70AFFF):
		return "Nepal"
	case icao >= uint32(0x70C000) && icao <= uint32(0x70C3FF):
		return "Oman"
	case icao >= uint32(0x70E000) && icao <= uint32(0x70EFFF):
		return "Cambodia"
	case icao >= uint32(0x710000) && icao <= uint32(0x717FFF):
		return "Saudi Arabia"
	case icao >= uint32(0x718000) && icao <= uint32(0x71FFFF):
		return "Korea (South)"
	case icao >= uint32(0x720000) && icao <= uint32(0x727FFF):
		return "Korea (North)"
	case icao >= uint32(0x728000) && icao <= uint32(0x72FFFF):
		return "Iraq"
	case icao >= uint32(0x730000) && icao <= uint32(0x737FFF):
		return "Iran"
	case icao >= uint32(0x738000) && icao <= uint32(0x73FFFF):
		return "Israel"
	case icao >= uint32(0x740000) && icao <= uint32(0x747FFF):
		return "Jordan"
	case icao >= uint32(0x748000) && icao <= uint32(0x74FFFF):
		return "Lebanon"
	case icao >= uint32(0x750000) && icao <= uint32(0x757FFF):
		return "Malaysia"
	case icao >= uint32(0x758000) && icao <= uint32(0x75FFFF):
		return "Philippines"
	case icao >= uint32(0x760000) && icao <= uint32(0x767FFF):
		return "Pakistan"
	case icao >= uint32(0x768000) && icao <= uint32(0x76FFFF):
		return "Singapore"
	case icao >= uint32(0x770000) && icao <= uint32(0x777FFF):
		return "Sri Lanka"
	case icao >= uint32(0x778000) && icao <= uint32(0x77FFFF):
		return "Syria"
	case icao >= uint32(0x780000) && icao <= uint32(0x7BFFFF):
		return "China"
	case icao >= uint32(0x7C0000) && icao <= uint32(0x7FFFFF):
		return "Australia"
	case icao >= uint32(0x800000) && icao <= uint32(0x83FFFF):
		return "India"
	case icao >= uint32(0x840000) && icao <= uint32(0x87FFFF):
		return "Japan"
	case icao >= uint32(0x880000) && icao <= uint32(0x887FFF):
		return "Thailand"
	case icao >= uint32(0x888000) && icao <= uint32(0x88FFFF):
		return "Viet Nam"
	case icao >= uint32(0x890000) && icao <= uint32(0x890FFF):
		return "Yemen"
	case icao >= uint32(0x894000) && icao <= uint32(0x894FFF):
		return "Bahrain"
	case icao >= uint32(0x895000) && icao <= uint32(0x8953FF):
		return "Brunei"
	case icao >= uint32(0x896000) && icao <= uint32(0x896FFF):
		return "United Arab Emirates"
	case icao >= uint32(0x897000) && icao <= uint32(0x8973FF):
		return "Solomon Islands"
	case icao >= uint32(0x898000) && icao <= uint32(0x898FFF):
		return "Papua New Guinea"
	case icao >= uint32(0x899000) && icao <= uint32(0x8993FF):
		return "Taiwan (unofficial)"
	case icao >= uint32(0x8A0000) && icao <= uint32(0x8A7FFF):
		return "Indonesia"
	case icao >= uint32(0x900000) && icao <= uint32(0x9FFFFF):
		return "(reserved, NAM/PAC)"
	case icao >= uint32(0x900000) && icao <= uint32(0x9003FF):
		return "Marshall Islands"
	case icao >= uint32(0x901000) && icao <= uint32(0x9013FF):
		return "Cook Islands"
	case icao >= uint32(0x902000) && icao <= uint32(0x9023FF):
		return "Samoa"
	case icao >= uint32(0xA00000) && icao <= uint32(0xADF7C7):
		return "United States"
	case icao >= uint32(0xADF7C8) && icao <= uint32(0xAFFFFF):
		return "United States (Mil)"
	case icao >= uint32(0xB00000) && icao <= uint32(0xBFFFFF):
		return "(reserved)"
	case icao >= uint32(0xC00000) && icao <= uint32(0xC3FFFF):
		return "Canada"
	case icao >= uint32(0xC80000) && icao <= uint32(0xC87FFF):
		return "New Zealand"
	case icao >= uint32(0xC88000) && icao <= uint32(0xC88FFF):
		return "Fiji"
	case icao >= uint32(0xC8A000) && icao <= uint32(0xC8A3FF):
		return "Nauru"
	case icao >= uint32(0xC8C000) && icao <= uint32(0xC8C3FF):
		return "Saint Lucia"
	case icao >= uint32(0xC8D000) && icao <= uint32(0xC8D3FF):
		return "Tonga"
	case icao >= uint32(0xC8E000) && icao <= uint32(0xC8E3FF):
		return "Kiribati"
	case icao >= uint32(0xC90000) && icao <= uint32(0xC903FF):
		return "Vanuatu"
	case icao >= uint32(0xD00000) && icao <= uint32(0xDFFFFF):
		return "(reserved)"
	case icao >= uint32(0xE00000) && icao <= uint32(0xE3FFFF):
		return "Argentina"
	case icao >= uint32(0xE40000) && icao <= uint32(0xE7FFFF):
		return "Brazil"
	case icao >= uint32(0xE80000) && icao <= uint32(0xE80FFF):
		return "Chile"
	case icao >= uint32(0xE84000) && icao <= uint32(0xE84FFF):
		return "Ecuador"
	case icao >= uint32(0xE88000) && icao <= uint32(0xE88FFF):
		return "Paraguay"
	case icao >= uint32(0xE8C000) && icao <= uint32(0xE8CFFF):
		return "Peru"
	case icao >= uint32(0xE90000) && icao <= uint32(0xE90FFF):
		return "Uruguay"
	case icao >= uint32(0xE94000) && icao <= uint32(0xE94FFF):
		return "Bolivia"
	case icao >= uint32(0xEC0000) && icao <= uint32(0xEFFFFF):
		return "(reserved, CAR)"
	case icao >= uint32(0xF00000) && icao <= uint32(0xF07FFF):
		return "ICAO (1)"
	case icao >= uint32(0xF00000) && icao <= uint32(0xFFFFFF):
		return "(reserved)"
	case icao >= uint32(0xF09000) && icao <= uint32(0xF093FF):
		return "ICAO (2)"
	default:
		return ""
	}
}
	//From  	To    	Country  - Extracted from http://www.aerotransport.org/html/ICAO_hex_decode.html
/*	icaoRange := [][3]uint32 {
		{uint32(0x000000), uint32(0x003FFF)},
	}
*/


// Decode reverses the encode operation on a byte slice input
func DecodeMsgPack(buf []byte, out interface{}) error {
	r := bytes.NewBuffer(buf)
	hd := codec.MsgpackHandle{}
	dec := codec.NewDecoder(r, &hd)
	return dec.Decode(out)
}

// Encode writes an encoded object to a new bytes buffer
func EncodeMsgPack(in interface{}) (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)
	hd := codec.MsgpackHandle{}
	enc := codec.NewEncoder(buf, &hd)
	err := enc.Encode(in)
	return buf, err
}
