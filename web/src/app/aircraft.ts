export class EventSourceResponse {
  data: AircraftResponse
}

export class AircraftResponse {
  now: number;
  total:number;
  good:number;
  bad:number;
  modea:number;
  modesshort:number;
  modeslong:number;
  aircraft: Aircraft[];

  constructor(data: any) {
    Object.assign(this, data);
  }
}

export class Aircraft {
  icao: string;// uint32
  //
  call: string; // string
  xpdr: number;   // uint32
  //
//  ERawLat   //uint32
//  ERawLon   //uint32
//  ORawLat   //uint32
//  ORawLon   //uint32

//  Latitude  float64
//  Longitude float64
  alt: number; //int32
//  AltUnit   uint

  //ewd            uint  // 0 = East, 1 = West.
  //ewv            int32 // E/W velocity.
  //nsd            uint  // 0 = North, 1 = South.
  //nsv            int32 // N/S velocity.
//  VertRateSource uint  // Vertical rate source.
//  VertRateSign   uint  // Vertical rate sign.
//  VertRate       int32 // Vertical rate.
  spd: number;         // int32
  hdg: number;       // int32
//  HeadingIsValid bool

//  LastPing time.Time
//  LastPos  time.Time

  rssi: number; // float64

//  Mlat    bool
//  IsValid bool

}

