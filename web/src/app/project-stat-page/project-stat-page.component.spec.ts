import {ComponentFixture, TestBed} from "@angular/core/testing";
import {HttpClientTestingModule, HttpTestingController} from "@angular/common/http/testing";
import {RouterTestingModule} from "@angular/router/testing";

import {ProjectStatPageComponent} from "./project-stat-page.component";
import {DatabaseProjectServices} from "../services/database-project.services";

describe("ProjectStatPageComponent", () => {
  let component: ProjectStatPageComponent;
  let fixture: ComponentFixture<ProjectStatPageComponent>;
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ProjectStatPageComponent],
      imports: [HttpClientTestingModule, RouterTestingModule],
      providers: [DatabaseProjectServices],
    })
      .overrideComponent(ProjectStatPageComponent, {set: {template: ""}})
      .compileComponents();

    fixture = TestBed.createComponent(ProjectStatPageComponent);
    component = fixture.componentInstance;
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => httpMock.verify());

  it("should create (no query params -> empty selection)", () => {
    expect(component).toBeTruthy();
    expect(component.projects).toEqual([]);
    expect(component.ids).toEqual([]);
  });
});
