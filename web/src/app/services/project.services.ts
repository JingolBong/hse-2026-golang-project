import {Injectable} from "@angular/core";
import {HttpClient} from "@angular/common/http";
import {Observable} from "rxjs";
import {IRequest} from "../models/request.model";
import {IRequestObject} from "../models/requestObj.model";
import {ConfigurationService} from "./configuration.services";

@Injectable({
    providedIn: 'root'
})
export class ProjectServices {
  urlPath = ""

  constructor(private http: HttpClient, private configurationService: ConfigurationService) {
    this.urlPath = configurationService.getValue("pathUrl")
  }

  getAll(page: number, searchName: String): Observable<IRequest>{
    return this.http.get<IRequest>(
      `http://${this.urlPath}/projects?page=${page}&limit=10&search=${encodeURIComponent(searchName as string)}`
    );
  }

  addProject(key: String): Observable<IRequestObject>{
    return this.http.post<IRequestObject>(
      `http://${this.urlPath}/projects/${encodeURIComponent(key as string)}/update`, {}
    );
  }

  deleteProject(id: Number): Observable<IRequestObject> {
    return this.http.delete<IRequestObject>(`http://${this.urlPath}/projects/${id}`);
  }
}
