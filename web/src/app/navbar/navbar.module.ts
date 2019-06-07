import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { NavbarComponent } from './navbar.component';
import { NavbarSideComponent } from './navbar-side.component';
import { NavbarTopComponent } from './navbar-top.component';
import { FormsModule } from '@angular/forms';
import { CollapseModule } from 'ngx-bootstrap/collapse';
import { AppRoutingModule } from '../app-routing.module';

@NgModule({
  declarations: [NavbarComponent, NavbarSideComponent, NavbarTopComponent],
  exports: [NavbarComponent, NavbarSideComponent, NavbarTopComponent],
  imports: [
    AppRoutingModule,
    CollapseModule.forRoot(),
    CommonModule,
    FormsModule
  ]
})
export class NavbarModule { }
