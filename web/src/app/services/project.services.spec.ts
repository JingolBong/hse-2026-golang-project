import {TestBed} from "@angular/core/testing";
import {HttpClientTestingModule, HttpTestingController} from "@angular/common/http/testing";

import {ProjectServices} from "./project.services";

describe("ProjectServices", () => {
  let service: ProjectServices;
  let httpMock: HttpTestingController;
  const base = "http://localhost:8000/api/v1";

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [ProjectServices],
    });
    service = TestBed.inject(ProjectServices);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => httpMock.verify());

  it("should be created", () => {
    expect(service).toBeTruthy();
  });

  it("getAll() GETs the connector projects with page/limit/search params", () => {
    const payload = {data: [{Key: "ABC"}], pageInfo: {currentPage: 2}};
    let received: any;
    service.getAll(2, "foo").subscribe(r => (received = r));

    const req = httpMock.expectOne(
      r => r.url.split("?")[0] === `${base}/connector/projects` || r.urlWithParams.startsWith(`${base}/connector/projects?`),
    );
    expect(req.request.method).toBe("GET");
    expect(req.request.urlWithParams).toContain("page=2");
    expect(req.request.urlWithParams).toContain("limit=10");
    expect(req.request.urlWithParams).toContain("search=foo");
    req.flush(payload);

    expect(received).toEqual(payload);
  });

  it("getAll() sends an empty search param when name is null/undefined", () => {
    service.getAll(1, null as any).subscribe();

    const req = httpMock.expectOne(r => r.url.split("?")[0] === `${base}/connector/projects`);
    expect(req.request.urlWithParams).toContain("search=");
    expect(req.request.urlWithParams).not.toContain("search=null");
    req.flush({});
  });

  it("addProject() POSTs updateProject with the project key as a query param", () => {
    service.addProject("XYZ").subscribe();

    const req = httpMock.expectOne(r => r.url.split("?")[0] === `${base}/connector/updateProject`);
    expect(req.request.method).toBe("POST");
    expect(req.request.urlWithParams).toContain("project=XYZ");
    expect(req.request.body).toEqual({});
    req.flush({status: true});
  });

  it("deleteProject() DELETEs the project by id", () => {
    service.deleteProject(42).subscribe();

    const req = httpMock.expectOne(`${base}/projects/42`);
    expect(req.request.method).toBe("DELETE");
    req.flush({status: true});
  });
});
