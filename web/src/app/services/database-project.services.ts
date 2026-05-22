import {Injectable} from "@angular/core";
import {HttpClient} from "@angular/common/http";
import {Observable} from "rxjs";
import {IRequest} from "../models/request.model";
import {IRequestObject} from "../models/requestObj.model";
import {ConfigurationService} from "./configuration.services";

@Injectable({
  providedIn: 'root'
})
export class DatabaseProjectServices {
  urlPath = ""

  constructor(private http: HttpClient, private configurationService: ConfigurationService) {
    this.urlPath = configurationService.getValue("pathUrl")
  }

  getAll(): Observable<IRequest>{
    return this.http.get<IRequest>(`http://${this.urlPath}/myprojects`);
  }

  getProjectStatByID(id: string): Observable<IRequestObject> {
    return this.http.get<IRequestObject>(`http://${this.urlPath}/myprojects/${id}/stat`);
  }

  getComplitedGraph(taskNumber: string, projectName: Array<string>): Observable<IRequestObject> {
    const projects = projectName.map(p => 'projects=' + encodeURIComponent(p)).join('&');
    return this.http.get<IRequestObject>(
      `http://${this.urlPath}/graph/compare?task=${taskNumber}&${projects}`
    );
  }

  getGraph(taskNumber: string, projectName: string): Observable<IRequestObject> {
    return this.http.get<IRequestObject>(
      `http://${this.urlPath}/graph?task=${taskNumber}&project=${encodeURIComponent(projectName)}`
    );
  }

  makeGraph(taskNumber: string, projectName: string): Observable<IRequestObject> {
    return this.http.post<IRequestObject>(
      `http://${this.urlPath}/graph?task=${taskNumber}&project=${encodeURIComponent(projectName)}`, {}
    );
  }

  deleteGraphs(projectName: string): Observable<IRequestObject> {
    return this.http.delete<IRequestObject>(
      `http://${this.urlPath}/graphs?project=${encodeURIComponent(projectName)}`
    );
  }

  isAnalyzed(projectName: string): Observable<IRequestObject>{
    return this.http.get<IRequestObject>(
      `http://${this.urlPath}/isAnalyzed?project=${encodeURIComponent(projectName)}`
    );
  }

  isEmpty(projectName: string): Observable<IRequestObject>{
    return this.http.get<IRequestObject>(
      `http://${this.urlPath}/isEmpty?project=${encodeURIComponent(projectName)}`
    );
  }
}
