import { TestBed } from '@angular/core/testing';

import { AircraftService } from './aircraft.service';

describe('AircraftService', () => {
  beforeEach(() => TestBed.configureTestingModule({}));

  it('should be created', () => {
    const service: AircraftService = TestBed.get(AircraftService);
    expect(service).toBeTruthy();
  });
});
