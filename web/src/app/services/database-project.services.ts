import {Injectable} from "@angular/core";
import {HttpClient} from "@angular/common/http";
import {Observable} from "rxjs";
import {IRequest} from "../models/request.model";
import {IRequestObject} from "../models/requestObj.model";
import {ConfigurationService} from "./configuration.services";

@Injectable({
  providedIn: "root",
})
export class DatabaseProjectServices {
  private readonly apiBase: string;

  constructor(private http: HttpClient, configurationService: ConfigurationService) {
    const host = configurationService.getValue<string>("host", "localhost");
    const port = configurationService.getValue<number>("port", 8000);
    this.apiBase = `http://${host}:${port}/api/v1`;
  }

  getAll(): Observable<IRequest> {
    return this.http.get<IRequest>(`${this.apiBase}/projects`);
  }

  getProjectStatByID(id: string): Observable<IRequestObject> {
    return this.http.get<IRequestObject>(`${this.apiBase}/projects/${id}`);
  }

  getIssuesByProject(projectKey: string): Observable<IRequest> {
    const params = new URLSearchParams({project: projectKey});
    return this.http.get<IRequest>(`${this.apiBase}/issues?${params.toString()}`);
  }

  getComplitedGraph(taskNumber: string, projectName: Array<string>): Observable<IRequestObject> {
    const params = new URLSearchParams({project: projectName.join(",")});
    return this.http.get<IRequestObject>(
      `${this.apiBase}/compare/${taskNumber}?${params.toString()}`,
    );
  }

  getGraph(taskNumber: string, projectName: string): Observable<IRequestObject> {
    const params = new URLSearchParams({project: projectName});
    return this.http.get<IRequestObject>(
      `${this.apiBase}/graph/get/${taskNumber}?${params.toString()}`,
    );
  }

  makeGraph(taskNumber: string, projectName: string): Observable<IRequestObject> {
    const params = new URLSearchParams({project: projectName});
    return this.http.post<IRequestObject>(
      `${this.apiBase}/graph/make/${taskNumber}?${params.toString()}`,
      {},
    );
  }

  deleteGraphs(projectName: string): Observable<IRequestObject> {
    const params = new URLSearchParams({project: projectName});
    return this.http.delete<IRequestObject>(`${this.apiBase}/graph/delete?${params.toString()}`);
  }

  isAnalyzed(projectName: string): Observable<IRequestObject> {
    const params = new URLSearchParams({project: projectName});
    return this.http.get<IRequestObject>(`${this.apiBase}/isAnalyzed?${params.toString()}`);
  }
}
