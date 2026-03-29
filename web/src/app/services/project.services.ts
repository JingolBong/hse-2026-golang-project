import {Injectable} from "@angular/core";
import {HttpClient} from "@angular/common/http";
import {Observable, of} from "rxjs";
import {IRequest} from "../models/request.model";
import {IRequestObject} from "../models/requestObj.model";
import {ConfigurationService} from "./configuration.services";
import {getMockAllProjects, MOCK_ADD_PROJECT, MOCK_DELETE_PROJECT} from "../mocks/projects.mock";

@Injectable({
    providedIn: 'root'
})
export class ProjectServices {
  urlPath = ""

  constructor(private http: HttpClient, private configurationService: ConfigurationService) {
    this.urlPath = configurationService.getValue("pathUrl")
  }

  getAll(page: number, searchName: String): Observable<IRequest>{
    // TODO Написать запрос на получение всех проектов, учесть пагинацию, поиск
    // return this.http.get<IRequest>(`http://${this.urlPath}/projects?page=${page}&limit=10&search=${searchName}`);
    return of(getMockAllProjects(page, searchName as string));
  }

  addProject(key: String): Observable<IRequestObject>{
    // TODO Написать запрос на добавление проекта в БД. Добавление происходит по ключу проекта
    // return this.http.get<IRequestObject>(`http://${this.urlPath}/updateProject?project=${key}`);
    return of(MOCK_ADD_PROJECT);
  }

  deleteProject(id: Number): Observable<IRequestObject> {
    // TODO Написать запрос на удаление проекта. Удаление происходит по id проекта в БД.
    // return this.http.delete<IRequestObject>(`http://${this.urlPath}/projects/${id}`);
    return of(MOCK_DELETE_PROJECT);
  }
}
