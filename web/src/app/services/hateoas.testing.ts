import {Observable, of} from "rxjs";
import {HateoasService} from "./hateoas.service";

// Shared test double for HateoasService: resolves every relation synchronously
// to its concrete URL, so component/service specs don't have to flush the
// service-discovery request. Mirrors the real ServiceLinks the backend serves.
const base = "http://localhost:8000/api/v1";

const LINKS: Record<string, string> = {
  projects: `${base}/projects`,
  projectStat: `${base}/projects/{id}`,
  connectorProjects: `${base}/connector/projects`,
  connectorUpdate: `${base}/connector/updateProject`,
  issues: `${base}/issues`,
  compare: `${base}/compare/{task}`,
  graphGet: `${base}/graph/get/{task}`,
  graphMake: `${base}/graph/make/{task}`,
  graphDelete: `${base}/graph/delete`,
  isAnalyzed: `${base}/isAnalyzed`,
};

export class FakeHateoasService {
  resolve(rel: string): Observable<string> {
    return of(LINKS[rel]);
  }

  resolveTemplate(rel: string, params: Record<string, string | number>): Observable<string> {
    let url = LINKS[rel];
    for (const [key, value] of Object.entries(params)) {
      url = url.replace(`{${key}}`, encodeURIComponent(String(value)));
    }
    return of(url);
  }
}

export const provideFakeHateoas = {provide: HateoasService, useClass: FakeHateoasService};
