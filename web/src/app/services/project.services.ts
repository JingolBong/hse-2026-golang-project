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
export class ProjectServices {
  constructor(private http: HttpClient, private hateoas: HateoasService) {}

  getAll(page: number, searchName: String): Observable<IRequest> {
    const params = new URLSearchParams({
      page: String(page),
      limit: "10",
      search: String(searchName ?? ""),
    });
    return this.hateoas
      .resolve("connectorProjects")
      .pipe(switchMap(href => this.http.get<IRequest>(`${href}?${params.toString()}`)));
  }

  addProject(key: String): Observable<IRequestObject> {
    const params = new URLSearchParams({project: String(key)});
    return this.hateoas
      .resolve("connectorUpdate")
      .pipe(switchMap(href => this.http.post<IRequestObject>(`${href}?${params.toString()}`, {})));
  }

  // id is kept as a string: the backend serializes the project id as a JSON
  // string to preserve int64 precision, which overflows a JS number.
  deleteProject(id: string | number): Observable<IRequestObject> {
    return this.hateoas
      .resolveTemplate("projectStat", {id: String(id)})
      .pipe(switchMap(href => this.http.delete<IRequestObject>(href)));
  }
}
