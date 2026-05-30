import {TestBed} from "@angular/core/testing";
import {HttpClientTestingModule, HttpTestingController} from "@angular/common/http/testing";

import {ConfigurationService} from "./configuration.services";

describe("ConfigurationService", () => {
  let service: ConfigurationService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [ConfigurationService],
    });
    service = TestBed.inject(ConfigurationService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => httpMock.verify());

  it("should be created", () => {
    expect(service).toBeTruthy();
  });

  it("load() requests the yaml config as text and parses it", () => {
    let completed = false;
    service.load().subscribe(() => (completed = true));

    const req = httpMock.expectOne("/assets/config.yaml");
    expect(req.request.method).toBe("GET");
    expect(req.request.responseType).toBe("text");
    req.flush("host: example.com\nport: 9090\n");

    expect(completed).toBeTrue();
    expect(service.getValue("host")).toBe("example.com");
    expect(service.getValue("port")).toBe(9090);
  });

  it("getValue returns the default when the key is absent", () => {
    service.load().subscribe();
    httpMock.expectOne("/assets/config.yaml").flush("host: only-host\n");

    expect(service.getValue("missing", "fallback")).toBe("fallback");
    expect(service.getValue("port", 8000)).toBe(8000);
  });

  it("getValue returns the default before any config is loaded", () => {
    expect(service.getValue("host", "localhost")).toBe("localhost");
  });

  it("treats an empty/null yaml document as an empty config", () => {
    service.load().subscribe();
    httpMock.expectOne("/assets/config.yaml").flush("");

    expect(service.getValue("host", "localhost")).toBe("localhost");
  });
});
