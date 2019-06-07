import {
  Component,
  OnInit,
  TemplateRef,
  ViewChild,
  ViewEncapsulation
} from '@angular/core';

import {Aircraft, Metrics} from '../aircraft';
import {AircraftService} from "../aircraft.service";
import {NgxDataTableConfig, TableConfig, TableEvent} from "patternfly-ng";

@Component({
  encapsulation: ViewEncapsulation.None,
  selector: 'app-aircraft',
  templateUrl: './aircraft.component.html',
  styleUrls: ['./aircraft.component.less']
})
export class AircraftComponent implements OnInit {
  @ViewChild('icaoTemplate') icaoTemplate: TemplateRef<any>;
  @ViewChild('callTemplate') callTemplate: TemplateRef<any>;
  @ViewChild('rngTemplate') rngTemplate: TemplateRef<any>;
  @ViewChild('altTemplate') altTemplate: TemplateRef<any>;
  @ViewChild('hdgTemplate') hdgTemplate: TemplateRef<any>;
  @ViewChild('xpdrTemplate') xpdrTemplate: TemplateRef<any>;
  @ViewChild('spdTemplate') spdTemplate: TemplateRef<any>;

  aircraft: Aircraft[];
  selectedAircraft: Aircraft;
  metrics: Metrics;

  subscription;

  actionsText: string = '';
  columns: any[];
  rows: any[];
  tableConfig: TableConfig;
  dataTableConfig: NgxDataTableConfig;

  onSelect(ac: Aircraft): void {
    this.selectedAircraft = ac;
  }

  constructor(private aircraftService: AircraftService) {
  }

  ngOnInit(): void {
    this.columns = [{
      cellTemplate: this.icaoTemplate,
      draggable: true,
      prop: 'icao',
      name: 'ICAO',
      resizeable: true
    }, {
      cellTemplate: this.callTemplate,
      draggable: true,
      prop: 'call',
      name: 'Call',
      resizeable: true
    }, {
      cellTemplate: this.rngTemplate,
      comparator: this.rangeComparator.bind(this),
      draggable: true,
      prop: 'rng',
      name: 'Range (nm)',
      resizeable: true
    }, {
      cellTemplate: this.hdgTemplate,
      draggable: true,
      prop: 'hdg',
      name: 'Heading',
      resizeable: true
    }, {
      cellTemplate: this.altTemplate,
      draggable: true,
      prop: 'alt',
      name: 'Altitude',
      resizeable: true
    }, {
      cellTemplate: this.xpdrTemplate,
      draggable: true,
      prop: 'xpdr',
      name: 'Squawk',
      resizeable: true
    },  {
      cellTemplate: this.spdTemplate,
      draggable: true,
      prop: 'spd',
      name: 'Speed',
      resizeable: true
    },
    ];

    this.tableConfig = {
      showCheckbox: false,
    } as TableConfig;

    this.dataTableConfig = {
      selectionType: 'single',
    };

    this.getAircraft();

  }

  ngOnDestroy() {
    this.subscription.unsubscribe();
  }

  getAircraft(): void {
    this.subscription = this.aircraftService.getAircraftStream()
      .subscribe(event => {
        this.metrics = event.metrics;
        this.aircraft = event.aircraft;
        // this.rows = event.aircraft;
        // console.log(event.metrics);
        // console.log(event.aircraft);
      });
  }

  handleSelectionChange($event: TableEvent): void {
    console.log("Selection Change event");
    this.selectedAircraft = $event.row;
    this.actionsText = $event.selectedRows.length + ' rows selected\r\n' + this.actionsText;
  }

  rangeComparator(propA, propB) {
    if (propA == '') {
      if (propB == '') {
        return 0; // they're both null
      }
      else {
        return 1; // only A is null
      }
    }
    else if (propB == '') {
      return -1; // only B is null
    }
    else if (propA < propB) {
      return -1;
    } else if (propA > propB) {
      return 1;
    }
  }

}
