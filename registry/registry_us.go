package registry

import (
	"strings"
)

const (
	US_RELEASABLE_AIRCRAFT_CURRENT         = "https://registry.faa.gov/database/ReleasableAircraft.zip"
	US_RELEASABLE_AIRCRAFT_YEARLY_TEMPLATE = "https://registry.faa.gov/database/yearly/ReleasableAircraft.%s.zip"
	US_REGISTRATION_FILENAME               = "US_Registrations.zip"
	YEAR                                   = "year"
)

type Master struct {
	//N-NUMBER,SERIAL NUMBER,MFR MDL CODE,ENG MFR MDL,YEAR MFR,TYPE REGISTRANT,NAME,STREET,STREET2,CITY,STATE,
	// ZIP CODE,REGION,COUNTY,COUNTRY,LAST ACTION DATE,CERT ISSUE DATE,CERTIFICATION,TYPE AIRCRAFT,TYPE ENGINE,
	// STATUS CODE,MODE S CODE,FRACT OWNER,AIR WORTH DATE,OTHER NAMES(1),OTHER NAMES(2),OTHER NAMES(3),
	// OTHER NAMES(4),OTHER NAMES(5),EXPIRATION DATE,UNIQUE ID,KIT MFR, KIT MODEL,MODE S CODE HEX,
	N_NUMBER      string `csv:"N-NUMBER,omitempty"`
	SERIAL_NUMBER string `csv:"SERIAL NUMBER,omitempty"`
	MFR_MDL_CODE  string `csv:"MFR MDL CODE,omitempty"`
	YEAR_MFR      string `csv:"YEAR MFR,omitempty"`
	NAME          string `csv:"NAME,omitempty"`
	TYPE_AIRCRAFT string `csv:"TYPE AIRCRAFT,omitempty"`
	TYPE_ENGINE   string `csv:"TYPE ENGINE,omitempty"`
	//MODE_S_CODE     string `csv:"MODE S CODE"`
	MODE_S_CODE_HEX string `csv:"MODE S CODE HEX" storm:"id"`
}

type AircraftRef struct {
	//CODE,MFR,MODEL,TYPE-ACFT,TYPE-ENG,AC-CAT,BUILD-CERT-IND,NO-ENG,NO-SEATS,AC-WEIGHT,SPEED,
	CODE      string `csv:"CODE"`
	MFR       string `csv:"MFR,omitempty"`
	MODEL     string `csv:"MODEL,omitempty"`
	TYPE_ACFT string `csv:"TYPE-ACFT,omitempty"`
}

func USRegistrationKey(value interface{}) string {
	return value.(*AircraftRef).CODE
}

func USMasterKey(value interface{}) string {
	return value.(*Master).MODE_S_CODE_HEX
}

func TrimSpace(field string, column string, v interface{}) string {
	return strings.TrimSpace(field)
}
