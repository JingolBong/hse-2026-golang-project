import {Injectable} from "@angular/core";
import {HttpClient} from "@angular/common/http";
import {Observable, of} from "rxjs";
import {IRequest} from "../models/request.model";
import {IRequestObject} from "../models/requestObj.model";
import {ConfigurationService} from "./configuration.services";
import {MOCK_MY_PROJECTS, getMockProjectStat, MOCK_IS_ANALYZED, MOCK_IS_EMPTY, MOCK_STATUS_OK} from "../mocks/projects.mock";
import {getMockGraph, getMockCompareGraph} from "../mocks/graphs.mock";

@Injectable({
  providedIn: 'root'
})
export class DatabaseProjectServices {
  urlPath = ""

  constructor(private http: HttpClient, private configurationService: ConfigurationService) {
    this.urlPath = configurationService.getValue("pathUrl")
  }

  getAll(): Observable<IRequest>{
    // TODO Написать запрос на получение всех проектов
    // return this.http.get<IRequest>(`http://${this.urlPath}/myprojects`);
    return of(MOCK_MY_PROJECTS);
  }

  getProjectStatByID(id: string): Observable<IRequestObject> {
    // TODO Написать запрос на получение статистики проекта по ID
    // return this.http.get<IRequestObject>(`http://${this.urlPath}/myprojects/${id}/stat`);
    return of(getMockProjectStat(id));
  }

  getComplitedGraph(taskNumber: string, projectName: Array<string>): Observable<IRequestObject> {
    // TODO Написать запрос на получение сравнения
    // return this.http.get<IRequestObject>(`http://${this.urlPath}/graph/compare?task=${taskNumber}&${projectName.map(p => 'projects=' + p).join('&')}`);
    return of(getMockCompareGraph(taskNumber, projectName));
  }

  getGraph(taskNumber: string, projectName: string): Observable<IRequestObject> {
    // TODO Написать запрос на получение графа
    // return this.http.get<IRequestObject>(`http://${this.urlPath}/graph?task=${taskNumber}&project=${projectName}`);
    return of(getMockGraph(taskNumber));
  }

  makeGraph(taskNumber: string, projectName: string): Observable<IRequestObject> {
    // TODO Написать запрос на создание графа
    // return this.http.post<IRequestObject>(`http://${this.urlPath}/graph?task=${taskNumber}&project=${projectName}`, {});
    return of(MOCK_STATUS_OK);
  }

  deleteGraphs(projectName: string): Observable<IRequestObject> {
    // TODO Написать запрос на удаление графа
    // return this.http.delete<IRequestObject>(`http://${this.urlPath}/graphs?project=${projectName}`);
    return of(MOCK_STATUS_OK);
  }

  isAnalyzed(projectName: string): Observable<IRequestObject>{
    // TODO Написать запрос
    // return this.http.get<IRequestObject>(`http://${this.urlPath}/isAnalyzed?project=${projectName}`);
    return of(MOCK_IS_ANALYZED);
  }

  isEmpty(projectName: string): Observable<IRequestObject>{
    // TODO Написать запрос
    // return this.http.get<IRequestObject>(`http://${this.urlPath}/isEmpty?project=${projectName}`);
    return of(MOCK_IS_EMPTY);
  }
}
