import { Component, OnInit } from '@angular/core';
import { Aircraft } from '../aircraft';
import { AircraftService } from "../aircraft.service";

@Component({
  selector: 'app-aircraft',
  templateUrl: './aircraft.component.html',
  styleUrls: ['./aircraft.component.css']
})
export class AircraftComponent implements OnInit {
  aircraft: Aircraft[];
  subscription;

  selectedAircraft: Aircraft;

  onSelect(ac: Aircraft): void {
    this.selectedAircraft = ac;
  }

  constructor(private aircraftService: AircraftService) { }

  ngOnInit() {
    this.getAircraft();
  }

  ngOnDestroy() {
    this.subscription.unsubscribe();
  }

  getAircraft(): void {
    this.subscription = this.aircraftService.getAircraftStream()
      .subscribe(event => {
        this.aircraft = event.aircraft;
        // console.log(event.aircraft);
      });
  }

}
