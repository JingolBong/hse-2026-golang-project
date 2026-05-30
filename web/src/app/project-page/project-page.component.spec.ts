import {ComponentFixture, TestBed} from "@angular/core/testing";
import {HttpClientTestingModule, HttpTestingController} from "@angular/common/http/testing";

import {ProjectPageComponent} from "./project-page.component";
import {ProjectServices} from "../services/project.services";

describe("ProjectPageComponent", () => {
  let component: ProjectPageComponent;
  let fixture: ComponentFixture<ProjectPageComponent>;
  let httpMock: HttpTestingController;
  const base = "http://localhost:8000/api/v1";

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ProjectPageComponent],
      imports: [HttpClientTestingModule],
      providers: [ProjectServices],
    })
      .overrideComponent(ProjectPageComponent, {set: {template: ""}})
      .compileComponents();

    fixture = TestBed.createComponent(ProjectPageComponent);
    component = fixture.componentInstance;
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => httpMock.verify());

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  it("ngOnInit loads the first page and clears the loading flag", () => {
    fixture.detectChanges();

    const req = httpMock.expectOne(r => r.url.split("?")[0] === `${base}/connector/projects`);
    expect(req.request.urlWithParams).toContain("page=1");
    req.flush({data: [{Key: "ABC"}], pageInfo: {currentPage: 1, projectsCount: 1}});

    expect(component.loading).toBeFalse();
    expect(component.projects.length).toBe(1);
    expect(component.pageInfo.currentPage).toBe(1);
  });

  it("gty() reloads the requested page", () => {
    component.gty(3);

    const req = httpMock.expectOne(r => r.url.split("?")[0] === `${base}/connector/projects`);
    expect(req.request.urlWithParams).toContain("page=3");
    req.flush({data: [], pageInfo: {currentPage: 3, projectsCount: 0}});

    expect(component.pageInfo.currentPage).toBe(3);
    expect(component.loading).toBeFalse();
  });
});
