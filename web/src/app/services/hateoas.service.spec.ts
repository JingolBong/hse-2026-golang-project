import {TestBed} from "@angular/core/testing";
import {HttpClientTestingModule, HttpTestingController} from "@angular/common/http/testing";

import {HateoasService} from "./hateoas.service";

describe("HateoasService", () => {
  let service: HateoasService;
  let httpMock: HttpTestingController;
  const discovery = "http://localhost:8000/api/v1/resource/services";

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [HateoasService],
    });
    service = TestBed.inject(HateoasService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => httpMock.verify());

  it("should be created", () => {
    expect(service).toBeTruthy();
  });

  it("baseUrl() exposes the configured root", () => {
    expect(service.baseUrl()).toBe("http://localhost:8000");
  });

  it("resolve() fetches service discovery and returns the matching href", () => {
    let href: string | undefined;
    service.resolve("projects").subscribe(h => (href = h));

    const req = httpMock.expectOne(discovery);
    expect(req.request.method).toBe("GET");
    req.flush({_links: {projects: {href: "http://localhost:8000/api/v1/projects"}}});

    expect(href).toBe("http://localhost:8000/api/v1/projects");
  });

  it("resolve() errors when the requested rel is missing", () => {
    let error: Error | undefined;
    service.resolve("nope").subscribe({
      next: () => fail("expected an error"),
      error: e => (error = e),
    });

    httpMock.expectOne(discovery).flush({_links: {other: {href: "x"}}});

    expect(error).toBeInstanceOf(Error);
    expect(error!.message).toContain('"nope"');
  });

  it("caches discovery so concurrent resolves trigger a single request", () => {
    let a: string | undefined;
    let b: string | undefined;
    service.resolve("a").subscribe(h => (a = h));
    service.resolve("b").subscribe(h => (b = h));

    const req = httpMock.expectOne(discovery);
    req.flush({_links: {a: {href: "/a"}, b: {href: "/b"}}});

    expect(a).toBe("/a");
    expect(b).toBe("/b");
    httpMock.expectNone(discovery);
  });

  it("tolerates a discovery response with no _links", () => {
    let error: Error | undefined;
    service.resolve("any").subscribe({error: e => (error = e)});

    httpMock.expectOne(discovery).flush({});

    expect(error).toBeInstanceOf(Error);
  });

  it("resolveTemplate() substitutes path params in the href", () => {
    let href: string | undefined;
    service.resolveTemplate("graphGet", {task: 3}).subscribe(h => (href = h));

    httpMock
      .expectOne(discovery)
      .flush({_links: {graphGet: {href: "http://localhost:8000/api/v1/graph/get/{task}"}}});

    expect(href).toBe("http://localhost:8000/api/v1/graph/get/3");
  });

  it("does not cache a failed discovery so the next call retries", () => {
    service.resolve("projects").subscribe({next: () => fail("expected an error"), error: () => {}});
    httpMock.expectOne(discovery).flush(
      {message: "boom"},
      {status: 500, statusText: "Server Error"},
    );

    // A second resolve must issue a fresh discovery request, not replay the error.
    let href: string | undefined;
    service.resolve("projects").subscribe(h => (href = h));
    httpMock
      .expectOne(discovery)
      .flush({_links: {projects: {href: "http://localhost:8000/api/v1/projects"}}});

    expect(href).toBe("http://localhost:8000/api/v1/projects");
  });
});
