import {Injectable} from "@angular/core";
import {HttpClient} from "@angular/common/http";
import {Observable} from "rxjs";
import {IRequest} from "../models/request.model";
import {IRequestObject} from "../models/requestObj.model";
import {ConfigurationService} from "./configuration.services";

@Injectable({
  providedIn: "root",
})
export class ProjectServices {
  private readonly apiBase: string;

  constructor(private http: HttpClient, configurationService: ConfigurationService) {
    const host = configurationService.getValue<string>("host", "localhost");
    const port = configurationService.getValue<number>("port", 8000);
    this.apiBase = `http://${host}:${port}/api/v1`;
  }

  getAll(page: number, searchName: String): Observable<IRequest> {
    const params = new URLSearchParams({
      page: String(page),
      limit: "10",
      search: String(searchName ?? ""),
    });
    return this.http.get<IRequest>(`${this.apiBase}/connector/projects?${params.toString()}`);
  }

  addProject(key: String): Observable<IRequestObject> {
    const params = new URLSearchParams({project: String(key)});
    return this.http.post<IRequestObject>(
      `${this.apiBase}/connector/updateProject?${params.toString()}`,
      {},
    );
  }

  deleteProject(id: Number): Observable<IRequestObject> {
    return this.http.delete<IRequestObject>(`${this.apiBase}/projects/${id}`);
  }
}
