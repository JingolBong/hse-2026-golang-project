import {ComponentFixture, TestBed} from "@angular/core/testing";
import {HttpClientTestingModule, HttpTestingController} from "@angular/common/http/testing";

import {MyProjectPageComponent} from "./my-project-page.component";
import {DatabaseProjectServices} from "../services/database-project.services";

describe("MyProjectPageComponent", () => {
  let component: MyProjectPageComponent;
  let fixture: ComponentFixture<MyProjectPageComponent>;
  let httpMock: HttpTestingController;
  const base = "http://localhost:8000/api/v1";

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [MyProjectPageComponent],
      imports: [HttpClientTestingModule],
      providers: [DatabaseProjectServices],
    })
      .overrideComponent(MyProjectPageComponent, {set: {template: ""}})
      .compileComponents();

    fixture = TestBed.createComponent(MyProjectPageComponent);
    component = fixture.componentInstance;
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => httpMock.verify());

  it("should create", () => {
    expect(component).toBeTruthy();
  });

  it("ngOnInit loads stored projects and marks the list as inited", () => {
    fixture.detectChanges();

    const req = httpMock.expectOne(`${base}/projects`);
    expect(req.request.method).toBe("GET");
    req.flush({data: [{Key: "ABC"}]});

    expect(component.myProjects.length).toBe(1);
    expect(component.loading).toBeFalse();
    expect(component.inited).toBeTrue();
    expect(component.noProjects).toBeFalse();
  });

  it("flags noProjects when the backend returns an empty list", () => {
    fixture.detectChanges();

    httpMock.expectOne(`${base}/projects`).flush({data: []});

    expect(component.noProjects).toBeTrue();
    expect(component.myProjects.length).toBe(0);
  });

  it("childOnChecked tracks and untracks checked projects", () => {
    component.childOnChecked({Name: "ABC", Id: 1, Checked: true} as any);
    expect(component.checked.get("ABC")).toBe(1);

    component.childOnChecked({Name: "ABC", Id: 1, Checked: false} as any);
    expect(component.checked.has("ABC")).toBeFalse();
  });
});
