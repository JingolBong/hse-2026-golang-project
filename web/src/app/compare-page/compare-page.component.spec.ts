import {ComponentFixture, TestBed} from "@angular/core/testing";
import {HttpClientTestingModule, HttpTestingController} from "@angular/common/http/testing";
import {RouterTestingModule} from "@angular/router/testing";

import {ComparePageComponent} from "./compare-page.component";
import {DatabaseProjectServices} from "../services/database-project.services";

describe("ComparePageComponent", () => {
  let component: ComparePageComponent;
  let fixture: ComponentFixture<ComparePageComponent>;
  let httpMock: HttpTestingController;
  const base = "http://localhost:8000/api/v1";

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ComparePageComponent],
      imports: [HttpClientTestingModule, RouterTestingModule],
      providers: [DatabaseProjectServices],
    })
      .overrideComponent(ComparePageComponent, {set: {template: ""}})
      .compileComponents();

    fixture = TestBed.createComponent(ComparePageComponent);
    component = fixture.componentInstance;
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => httpMock.verify());

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  it("ngOnInit loads projects available for comparison", () => {
    fixture.detectChanges();

    const req = httpMock.expectOne(`${base}/projects`);
    expect(req.request.method).toBe("GET");
    req.flush({data: [{Key: "ABC"}, {Key: "XYZ"}]});

    expect(component.projects.length).toBe(2);
    expect(component.inited).toBeTrue();
    expect(component.noProjects).toBeFalse();
  });

  it("flags noProjects on an empty list", () => {
    fixture.detectChanges();
    httpMock.expectOne(`${base}/projects`).flush({data: []});
    expect(component.noProjects).toBeTrue();
  });
});
