import {ComponentFixture, TestBed} from "@angular/core/testing";
import {HttpClientTestingModule, HttpTestingController} from "@angular/common/http/testing";
import {RouterTestingModule} from "@angular/router/testing";
import {Router} from "@angular/router";

import {MyProjectComponent} from "./my-project.component";
import {DatabaseProjectServices} from "../../services/database-project.services";
import {provideFakeHateoas} from "../../services/hateoas.testing";

describe("MyProjectComponent", () => {
  let component: MyProjectComponent;
  let fixture: ComponentFixture<MyProjectComponent>;
  let httpMock: HttpTestingController;
  let router: Router;
  const base = "http://localhost:8000/api/v1";

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [MyProjectComponent],
      imports: [HttpClientTestingModule, RouterTestingModule],
      providers: [DatabaseProjectServices, provideFakeHateoas],
    })
      .overrideComponent(MyProjectComponent, {set: {template: ""}})
      .compileComponents();

    fixture = TestBed.createComponent(MyProjectComponent);
    component = fixture.componentInstance;
    component.myProject = {Existence: true, Id: 7, Key: "ABC", Name: "Alpha", Url: "u"} as any;
    httpMock = TestBed.inject(HttpTestingController);
    router = TestBed.inject(Router);
  });

  afterEach(() => httpMock.verify());

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  it("ngOnInit loads stat + analysis status and builds the checkbox list", () => {
    fixture.detectChanges();

    const statReq = httpMock.expectOne(`${base}/projects/7`);
    expect(statReq.request.method).toBe("GET");
    statReq.flush({data: {allIssuesCount: 10, openIssuesCount: 3, averageTime: 5}});

    const analyzedReq = httpMock.expectOne(r => r.url.split("?")[0] === `${base}/isAnalyzed`);
    expect(analyzedReq.request.urlWithParams).toContain("project=Alpha");
    analyzedReq.flush({data: {isAnalyzed: true}});

    expect(component.checkboxes.length).toBe(6);
    expect(component.stat.AllIssuesCount).toBe(10);
    expect(component.stat.OpenIssuesCount).toBe(3);
    expect(component.processed).toBeTrue();
  });

  it("childOnChecked toggles the matching checkbox state", () => {
    fixture.detectChanges();
    httpMock.expectOne(`${base}/projects/7`).flush({data: {}});
    httpMock.expectOne(r => r.url.split("?")[0] === `${base}/isAnalyzed`).flush({data: {}});

    component.childOnChecked({ProjectName: "Alpha", BoxId: 2, Checked: true} as any);
    expect(component.setting.get("Alpha")).toBe(2);
    expect(component.checkboxes[1].Checked).toBeTrue();

    component.childOnChecked({ProjectName: "Alpha", BoxId: 2, Checked: false} as any);
    expect(component.setting.has("Alpha")).toBeFalse();
    expect(component.checkboxes[1].Checked).toBeFalse();
  });

  it("checkResult navigates to /project-stat with selected box ids", () => {
    const navSpy = spyOn(router, "navigate");
    fixture.detectChanges();
    httpMock.expectOne(`${base}/projects/7`).flush({data: {}});
    httpMock.expectOne(r => r.url.split("?")[0] === `${base}/isAnalyzed`).flush({data: {}});

    component.checkboxes[0].Checked = true;
    component.checkboxes[2].Checked = true;
    component.checkResult();

    expect(navSpy).toHaveBeenCalledWith(["/project-stat"], {
      queryParams: {keys: "Alpha", value: [1, 3]},
    });
  });

  it("disableCheckResultButton requires processed and completed work", () => {
    component.processed = false;
    expect(component.disableCheckResultButton()).toBeTrue();

    component.processed = true;
    component.checked = 1;
    component.complited = 1;
    expect(component.disableCheckResultButton()).toBeFalse();
  });
});
