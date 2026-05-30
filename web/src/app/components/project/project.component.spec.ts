import {ComponentFixture, TestBed} from "@angular/core/testing";
import {HttpClientTestingModule, HttpTestingController} from "@angular/common/http/testing";

import {ProjectComponent} from "./project.component";
import {ProjectServices} from "../../services/project.services";

describe("ProjectComponent", () => {
  let component: ProjectComponent;
  let fixture: ComponentFixture<ProjectComponent>;
  let httpMock: HttpTestingController;
  const base = "http://localhost:8000/api/v1";

  function setup(existence: boolean) {
    fixture = TestBed.createComponent(ProjectComponent);
    component = fixture.componentInstance;
    component.project = {Existence: existence, Id: 5, Key: "ABC", Name: "Alpha", Url: "u"} as any;
    fixture.detectChanges(); 
  }

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ProjectComponent],
      imports: [HttpClientTestingModule],
      providers: [ProjectServices],
    })
      .overrideComponent(ProjectComponent, {set: {template: ""}})
      .compileComponents();
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => httpMock.verify());

  it("should create and mirror Existence into the adding flag", () => {
    setup(true);
    expect(component).toBeTruthy();
    expect(component.adding).toBeTrue();
  });

  it("addMyProject adds the project when it is not yet tracked", () => {
    setup(false);
    component.addMyProject(component.project);

    const req = httpMock.expectOne(r => r.url.split("?")[0] === `${base}/connector/updateProject`);
    expect(req.request.method).toBe("POST");
    expect(req.request.urlWithParams).toContain("project=ABC");
    req.flush({status: true});

    expect(component.adding).toBeTrue();
  });

  it("addMyProject deletes the project when it is already tracked", () => {
    setup(true);
    component.addMyProject(component.project);

    const req = httpMock.expectOne(`${base}/projects/5`);
    expect(req.request.method).toBe("DELETE");
    req.flush({status: true});

    expect(component.adding).toBeFalse();
  });
});
