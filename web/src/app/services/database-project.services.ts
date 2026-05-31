import {Injectable} from "@angular/core";
import {HttpClient} from "@angular/common/http";
import {Observable} from "rxjs";
import {switchMap} from "rxjs/operators";
import {IRequest} from "../models/request.model";
import {IRequestObject} from "../models/requestObj.model";
import {HateoasService} from "./hateoas.service";

@Injectable({
  providedIn: "root",
})
export class DatabaseProjectServices {
  constructor(private http: HttpClient, private hateoas: HateoasService) {}

  getAll(): Observable<IRequest> {
    return this.hateoas
      .resolve("projects")
      .pipe(switchMap(href => this.http.get<IRequest>(href)));
  }

  getProjectStatByID(id: string): Observable<IRequestObject> {
    return this.hateoas
      .resolveTemplate("projectStat", {id})
      .pipe(switchMap(href => this.http.get<IRequestObject>(href)));
  }

  getIssuesByProject(projectKey: string): Observable<IRequest> {
    const params = new URLSearchParams({project: projectKey});
    return this.hateoas
      .resolve("issues")
      .pipe(switchMap(href => this.http.get<IRequest>(`${href}?${params.toString()}`)));
  }

  getComplitedGraph(taskNumber: string, projectName: Array<string>): Observable<IRequestObject> {
    const params = new URLSearchParams({project: projectName.join(",")});
    return this.hateoas
      .resolveTemplate("compare", {task: taskNumber})
      .pipe(switchMap(href => this.http.get<IRequestObject>(`${href}?${params.toString()}`)));
  }

  getGraph(taskNumber: string, projectName: string): Observable<IRequestObject> {
    const params = new URLSearchParams({project: projectName});
    return this.hateoas
      .resolveTemplate("graphGet", {task: taskNumber})
      .pipe(switchMap(href => this.http.get<IRequestObject>(`${href}?${params.toString()}`)));
  }

  makeGraph(taskNumber: string, projectName: string): Observable<IRequestObject> {
    const params = new URLSearchParams({project: projectName});
    return this.hateoas
      .resolveTemplate("graphMake", {task: taskNumber})
      .pipe(switchMap(href => this.http.post<IRequestObject>(`${href}?${params.toString()}`, {})));
  }

  deleteGraphs(projectName: string): Observable<IRequestObject> {
    const params = new URLSearchParams({project: projectName});
    return this.hateoas
      .resolve("graphDelete")
      .pipe(switchMap(href => this.http.delete<IRequestObject>(`${href}?${params.toString()}`)));
  }

  isAnalyzed(projectName: string): Observable<IRequestObject> {
    const params = new URLSearchParams({project: projectName});
    return this.hateoas
      .resolve("isAnalyzed")
      .pipe(switchMap(href => this.http.get<IRequestObject>(`${href}?${params.toString()}`)));
  }
}
