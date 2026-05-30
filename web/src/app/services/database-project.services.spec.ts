import {TestBed} from "@angular/core/testing";
import {HttpClientTestingModule, HttpTestingController} from "@angular/common/http/testing";

import {DatabaseProjectServices} from "./database-project.services";

describe("DatabaseProjectServices", () => {
  let service: DatabaseProjectServices;
  let httpMock: HttpTestingController;
  const base = "http://localhost:8000/api/v1";

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [DatabaseProjectServices],
    });
    service = TestBed.inject(DatabaseProjectServices);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => httpMock.verify());

  it("should be created", () => {
    expect(service).toBeTruthy();
  });

  it("getAll() GETs the stored projects", () => {
    const payload = {data: [{Key: "ABC"}]};
    let received: any;
    service.getAll().subscribe(r => (received = r));

    const req = httpMock.expectOne(`${base}/projects`);
    expect(req.request.method).toBe("GET");
    req.flush(payload);
    expect(received).toEqual(payload);
  });

  it("getProjectStatByID() GETs a single project by id", () => {
    service.getProjectStatByID("7").subscribe();
    const req = httpMock.expectOne(`${base}/projects/7`);
    expect(req.request.method).toBe("GET");
    req.flush({});
  });

  it("getIssuesByProject() GETs issues filtered by project key", () => {
    service.getIssuesByProject("ABC").subscribe();
    const req = httpMock.expectOne(r => r.url.split("?")[0] === `${base}/issues`);
    expect(req.request.method).toBe("GET");
    expect(req.request.urlWithParams).toContain("project=ABC");
    req.flush({});
  });

  it("getComplitedGraph() joins multiple project names with a comma", () => {
    service.getComplitedGraph("3", ["A", "B", "C"]).subscribe();
    const req = httpMock.expectOne(r => r.url.split("?")[0] === `${base}/compare/3`);
    expect(req.request.method).toBe("GET");
    expect(decodeURIComponent(req.request.urlWithParams)).toContain("project=A,B,C");
    req.flush({});
  });

  it("getGraph() GETs a stored graph for a single project", () => {
    service.getGraph("5", "ABC").subscribe();
    const req = httpMock.expectOne(r => r.url.split("?")[0] === `${base}/graph/get/5`);
    expect(req.request.method).toBe("GET");
    expect(req.request.urlWithParams).toContain("project=ABC");
    req.flush({});
  });

  it("makeGraph() POSTs to build a graph", () => {
    service.makeGraph("5", "ABC").subscribe();
    const req = httpMock.expectOne(r => r.url.split("?")[0] === `${base}/graph/make/5`);
    expect(req.request.method).toBe("POST");
    expect(req.request.urlWithParams).toContain("project=ABC");
    expect(req.request.body).toEqual({});
    req.flush({});
  });

  it("deleteGraphs() DELETEs graphs for a project", () => {
    service.deleteGraphs("ABC").subscribe();
    const req = httpMock.expectOne(r => r.url.split("?")[0] === `${base}/graph/delete`);
    expect(req.request.method).toBe("DELETE");
    expect(req.request.urlWithParams).toContain("project=ABC");
    req.flush({});
  });

  it("isAnalyzed() GETs the analysis status for a project", () => {
    service.isAnalyzed("ABC").subscribe();
    const req = httpMock.expectOne(r => r.url.split("?")[0] === `${base}/isAnalyzed`);
    expect(req.request.method).toBe("GET");
    expect(req.request.urlWithParams).toContain("project=ABC");
    req.flush({});
  });
});
