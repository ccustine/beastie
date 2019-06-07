import {Component, Input, OnInit} from '@angular/core';
import {InfoStatusCardConfig} from "patternfly-ng";
import {Metrics} from "../aircraft";

@Component({
  selector: 'app-metrics',
  templateUrl: './metrics.component.html',
  styleUrls: ['./metrics.component.less']
})

export class MetricsComponent implements OnInit {
  @Input() metrics: Metrics;

  card1Config: InfoStatusCardConfig = {
    showTopBorder: true,
    htmlContent: true,
    title: 'TinyCore-local',
    href: '//www.redhat.com/',
    iconStyleClass: 'fa fa-shield',
    info: [
      'VM Name: ',
      'Host Name: localhost.localdomian',
      'IP Address: 10.9.62.100',
      'Power status: on'
    ]
  };

  constructor() { }

  ngOnInit() {
  }

}
