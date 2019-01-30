import { Component, OnInit, Input } from '@angular/core';
import { Aircraft} from "../aircraft";

@Component({
  selector: 'app-aircraft-detail',
  templateUrl: './aircraft-detail.component.html',
  styleUrls: ['./aircraft-detail.component.css']
})
export class AircraftDetailComponent implements OnInit {
  @Input() aircraft: Aircraft;

  constructor() { }

  ngOnInit() {
  }

}
