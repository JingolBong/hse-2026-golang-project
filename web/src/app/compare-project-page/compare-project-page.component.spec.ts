import {ComponentFixture, TestBed} from "@angular/core/testing";
import {HttpClientTestingModule, HttpTestingController} from "@angular/common/http/testing";
import {RouterTestingModule} from "@angular/router/testing";

import {CompareProjectPageComponent} from "./compare-project-page.component";
import {DatabaseProjectServices} from "../services/database-project.services";
import {provideFakeHateoas} from "../services/hateoas.testing";

describe("CompareProjectPageComponent", () => {
  let component: CompareProjectPageComponent;
  let fixture: ComponentFixture<CompareProjectPageComponent>;
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [CompareProjectPageComponent],
      imports: [HttpClientTestingModule, RouterTestingModule],
      providers: [DatabaseProjectServices, provideFakeHateoas],
    })
      .overrideComponent(CompareProjectPageComponent, {set: {template: ""}})
      .compileComponents();

    fixture = TestBed.createComponent(CompareProjectPageComponent);
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
