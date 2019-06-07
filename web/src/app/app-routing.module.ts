import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';

import { AircraftComponent } from '../app/aircraft/aircraft.component';

const routes: Routes = [{
  path: 'aircraft',
  component: AircraftComponent
},
];

@NgModule({
  imports: [RouterModule.forRoot(routes, { useHash: true })],
  exports: [RouterModule]
})
export class AppRoutingModule {}
