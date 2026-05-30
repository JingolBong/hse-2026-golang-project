import {ComponentFixture, TestBed} from "@angular/core/testing";
import {HttpClientTestingModule, HttpTestingController} from "@angular/common/http/testing";

import {IssuesPageComponent} from "./issues-page.component";
import {DatabaseProjectServices} from "../services/database-project.services";

describe("IssuesPageComponent", () => {
  let component: IssuesPageComponent;
  let fixture: ComponentFixture<IssuesPageComponent>;
  let httpMock: HttpTestingController;
  const base = "http://localhost:8000/api/v1";

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [IssuesPageComponent],
      imports: [HttpClientTestingModule],
      providers: [DatabaseProjectServices],
    })
      .overrideComponent(IssuesPageComponent, {set: {template: ""}})
      .compileComponents();

    fixture = TestBed.createComponent(IssuesPageComponent);
    component = fixture.componentInstance;
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => httpMock.verify());

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  it("ngOnInit loads the project list", () => {
    fixture.detectChanges();
    httpMock.expectOne(`${base}/projects`).flush({data: [{Key: "ABC"}]});

    expect(component.projects.length).toBe(1);
    expect(component.error).toBeNull();
  });

  it("ngOnInit sets an error message when project loading fails", () => {
    fixture.detectChanges();
    httpMock.expectOne(`${base}/projects`).flush("boom", {status: 500, statusText: "Server Error"});

    expect(component.error).toContain("список проектов");
  });

  it("selectProject loads issues for the chosen project and resets state", () => {
    fixture.detectChanges();
    httpMock.expectOne(`${base}/projects`).flush({data: []});

    component.selectProject("ABC");
    expect(component.selectedKey).toBe("ABC");
    expect(component.loading).toBeTrue();

    const req = httpMock.expectOne(r => r.url.split("?")[0] === `${base}/issues`);
    expect(req.request.urlWithParams).toContain("project=ABC");
    req.flush({data: [{key: "ABC-1"}, {key: "ABC-2"}]});

    expect(component.issues.length).toBe(2);
    expect(component.loaded).toBeTrue();
    expect(component.loading).toBeFalse();
  });

  it("selectProject sets an error message when issue loading fails", () => {
    fixture.detectChanges();
    httpMock.expectOne(`${base}/projects`).flush({data: []});

    component.selectProject("ABC");
    httpMock
      .expectOne(r => r.url.split("?")[0] === `${base}/issues`)
      .flush("boom", {status: 500, statusText: "Server Error"});

    expect(component.loading).toBeFalse();
    expect(component.error).toContain("задачи проекта");
  });
});
