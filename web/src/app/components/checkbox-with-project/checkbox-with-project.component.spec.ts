import {ComponentFixture, TestBed} from "@angular/core/testing";
import {HttpClientTestingModule} from "@angular/common/http/testing";

import {ProjectWithCheckboxComponent} from "./checkbox-with-project.component";
import {ProjectServices} from "../../services/project.services";
import {CheckedProject} from "../../models/check-element.model";

describe("ProjectWithCheckboxComponent", () => {
  let component: ProjectWithCheckboxComponent;
  let fixture: ComponentFixture<ProjectWithCheckboxComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ProjectWithCheckboxComponent],
      imports: [HttpClientTestingModule],
      providers: [ProjectServices],
    })
      .overrideComponent(ProjectWithCheckboxComponent, {set: {template: ""}})
      .compileComponents();

    fixture = TestBed.createComponent(ProjectWithCheckboxComponent);
    component = fixture.componentInstance;
    component.project = {Existence: true, Id: 9, Key: "ABC", Name: "Alpha", Url: "u"} as any;
  });

  it("should create and seed isChecked from project Existence", () => {
    fixture.detectChanges();
    expect(component).toBeTruthy();
    expect(component.isChecked).toBeTrue();
  });

  it("changed() emits a CheckedProject carrying the project name and id", () => {
    fixture.detectChanges();
    let emitted: CheckedProject | undefined;
    component.onChecked.subscribe((e: CheckedProject) => (emitted = e));

    component.isChecked = false;
    component.changed(false);

    expect(emitted).toEqual(new CheckedProject("Alpha", false, 9));
  });
});
