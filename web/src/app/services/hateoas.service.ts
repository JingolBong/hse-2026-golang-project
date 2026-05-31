import {Injectable} from "@angular/core";
import {HttpClient} from "@angular/common/http";
import {Observable, throwError} from "rxjs";
import {catchError, map, shareReplay} from "rxjs/operators";
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

  constructor(private http: HttpClient, private config: ConfigurationService) {
    const host = config.getValue("host", "localhost");
    const port = config.getValue("port", 8000);
    this.root = `http://${host}:${port}`;
  }

  /** Resolve a top-level service link by its relation name. */
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

  /**
   * Resolve a templated link and fill in its path params, e.g.
   * resolveTemplate("graphGet", {task: 3}) -> ".../api/v1/graph/get/3".
   */
  resolveTemplate(rel: string, params: Record<string, string | number>): Observable<string> {
    return this.resolve(rel).pipe(
      map(href => {
        let url = href;
        for (const [key, value] of Object.entries(params)) {
          url = url.replace(`{${key}}`, encodeURIComponent(String(value)));
        }
        return url;
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
          catchError(err => {
            // Don't let a transient discovery failure poison the cache: drop it
            // so the next call retries instead of replaying the error forever.
            this.linksCache$ = undefined;
            return throwError(() => err);
          }),
          shareReplay(1),
        );
    }
    return this.linksCache$;
  }
}
