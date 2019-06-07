import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { AppComponent } from './app.component';
import { AircraftComponent } from './aircraft/aircraft.component';
import { AircraftDetailComponent } from './aircraft-detail/aircraft-detail.component';
import { MessagesComponent } from './messages/messages.component';
import { HttpClientModule } from '@angular/common/http';
// import { TableModule } from 'patternfly-ng';
import { TableModule, NgxDataTableConfig, TableConfig, TableEvent } from 'patternfly-ng';
// NGX Bootstrap
import { BsDropdownConfig, BsDropdownModule } from 'ngx-bootstrap/dropdown';
// NGX Datatable
import { NgxDatatableModule } from '@swimlane/ngx-datatable';
import { NavbarModule } from './navbar/navbar.module';
import { MetricsComponent } from './metrics/metrics.component';
import { InfoStatusCardModule, InfoStatusCardConfig } from 'patternfly-ng';

@NgModule({
  declarations: [
    AppComponent,
    AircraftComponent,
    AircraftDetailComponent,
    MessagesComponent,
    MetricsComponent,
  ],
  imports: [
    BrowserModule,
    FormsModule,
    HttpClientModule,
    BsDropdownModule.forRoot(),
    NgxDatatableModule,
    TableModule,
    NavbarModule,
    InfoStatusCardModule,
  ],
  providers: [BsDropdownConfig],
  bootstrap: [AppComponent]
})
export class AppModule { }
