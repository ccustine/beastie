import {Injectable, NgZone} from '@angular/core';
import { AircraftResponse} from "./aircraft";
import { Observable, of } from 'rxjs';
import { MessageService } from './message.service';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { EventSourcePolyfill } from 'ng-event-source';

@Injectable({
  providedIn: 'root'
})
export class AircraftService {
  private aircraftUrl = '/aircraft';
  private streamUrl = '/stream?stream=aircraft';
  private zone = new NgZone({ enableLongStackTrace: false });
  private acr: AircraftResponse;

  constructor(private messageService: MessageService,
              private http: HttpClient) {

  }

  getAircraftStream(): Observable<AircraftResponse> {
    return Observable.create((observer) => {
      const eventSource = new EventSourcePolyfill(this.streamUrl, { heartbeatTimeout: 10000, connectionTimeout: 10000 });
      eventSource.onmessage = event => {
        // this.log("Firing observable " + acr.data)
        this.zone.run(() => {
          const json = JSON.parse(event.data);
          this.acr = new AircraftResponse(json);
          observer.next(this.acr)
        });
      }
      eventSource.onerror = error => this.zone.run(() => observer.error(error));
      return () => eventSource.close();
    });
  }

  getAircraft(): Observable<AircraftResponse> {
    this.log('AircraftService: fetching aircraft');
    return this.http.get<AircraftResponse>(this.aircraftUrl)
  }

  private log(message: string) {
    this.messageService.add(`HeroService: ${message}`);
  }

}
