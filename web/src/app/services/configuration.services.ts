import {Injectable} from "@angular/core";
import {HttpClient} from "@angular/common/http";
import {Observable, map} from "rxjs";
import {load as parseYaml} from "js-yaml";

@Injectable({
  providedIn: "root",
})
export class ConfigurationService {
  private configuration: Record<string, unknown> = {};

  constructor(private httpClient: HttpClient) {
  }

  load(): Observable<void> {
    return this.httpClient
      .get("/assets/config.yaml", {responseType: "text"})
      .pipe(
        map(raw => {
          const parsed = parseYaml(raw) as Record<string, unknown> | null;
          this.configuration = parsed ?? {};
        }),
      );
  }

  getValue<T = unknown>(key: string, defaultValue?: T): T {
    const value = this.configuration[key];
    return (value === undefined || value === null ? defaultValue : value) as T;
  }
}
