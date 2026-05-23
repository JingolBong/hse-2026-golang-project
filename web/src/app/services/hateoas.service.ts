import {Injectable} from "@angular/core";
import {HttpClient} from "@angular/common/http";
import {Observable, ReplaySubject, of} from "rxjs";
import {map, shareReplay, switchMap, tap} from "rxjs/operators";
import {ConfigurationService} from "./configuration.services";
import {Links} from "../models/links.model";

interface ServicesEnvelope {
  _links: Links;
  data: unknown;
  message?: string;
  status?: boolean;
}

@Injectable({providedIn: "root"})
export class HateoasService {
  private readonly root: string;
  private linksCache$?: Observable<Links>;
  private readonly ready$ = new ReplaySubject<Links>(1);

  constructor(private http: HttpClient, private config: ConfigurationService) {
    const host = config.getValue("host", "localhost");
    const port = config.getValue("port", 8000);
    this.root = `http://${host}:${port}`;
  }

  resolve(rel: string): Observable<string> {
    return this.links().pipe(
      map(links => {
        const link = links[rel];
        if (!link) {
          throw new Error(`HATEOAS link "${rel}" not found in service discovery`);
        }
        return link.href;
      }),
    );
  }

  baseUrl(): string {
    return this.root;
  }

  private links(): Observable<Links> {
    if (!this.linksCache$) {
      this.linksCache$ = this.http
        .get<ServicesEnvelope>(`${this.root}/api/v1/resource/services`)
        .pipe(
          map(envelope => envelope._links ?? {}),
          tap(links => this.ready$.next(links)),
          shareReplay(1),
        );
    }
    return this.linksCache$;
  }
}
